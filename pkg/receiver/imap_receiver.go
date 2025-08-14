package receiver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/domain"
	"github.com/headmail/headmail/pkg/repository"
)

// Receiver defines an interface for inbound mail receivers (bounce processors, webhooks, etc).
type Receiver interface {
	// Start starts background processing (returns when started or on error).
	Start(ctx context.Context) error
	// Stop requests graceful stop and waits for completion.
	Stop(ctx context.Context) error
}

// IMAPReceiver polls an IMAP mailbox and records bounce events.
type IMAPReceiver struct {
	cfg        config.SMTPConfig
	db         repository.DB
	mailbox    string
	pollPeriod time.Duration

	client    *imapclient.Client
	cancelCtx context.CancelFunc
	doneCh    chan struct{}
}

// NewIMAPReceiver creates a new IMAPReceiver.
// cfg is the SMTP.Receive section from application config.
func NewIMAPReceiver(cfg config.SMTPConfig, db repository.DB) *IMAPReceiver {
	period := time.Minute
	return &IMAPReceiver{
		cfg:        cfg,
		db:         db,
		mailbox:    cfg.Receive.Mailbox,
		pollPeriod: period,
		doneCh:     make(chan struct{}),
	}
}

// Start begins polling the IMAP mailbox in background.
func (r *IMAPReceiver) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	r.cancelCtx = cancel

	go func() {
		defer close(r.doneCh)
		ticker := time.NewTicker(r.pollPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := r.pollOnce(ctx); err != nil {
				log.Printf("imap receiver: poll error: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
	}()

	return nil
}

func (r *IMAPReceiver) Stop(ctx context.Context) error {
	if r.cancelCtx != nil {
		r.cancelCtx()
	}
	// wait for done
	select {
	case <-r.doneCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// dial connects to IMAP server (TLS if configured).
func (r *IMAPReceiver) dial() (*imapclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", r.cfg.Receive.Host, r.cfg.Receive.Port)
	if r.cfg.Receive.TLS {
		// DialTLS will do TLS
		c, err := imapclient.DialTLS(addr, nil)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	// plain TCP then STARTTLS if server supports it (not implemented here)
	c, err := imapclient.Dial(addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// pollOnce connects to IMAP, fetches UNSEEN messages and processes them.
func (r *IMAPReceiver) pollOnce(ctx context.Context) error {
	c, err := r.dial()
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Logout()
	}()

	// Login
	if err := c.Login(r.cfg.Receive.Username, r.cfg.Receive.Password); err != nil {
		return err
	}

	mbox := r.mailbox
	if mbox == "" {
		mbox = "INBOX"
	}
	_, err = c.Select(mbox, false)
	if err != nil {
		return err
	}

	// Search for UNSEEN messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(criteria)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)
	if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid, section.FetchItem()}, messages); err != nil {
		return err
	}

	var toMark imap.SeqSet
	for msg := range messages {
		if msg == nil {
			continue
		}
		rdr := msg.GetBody(section)
		if rdr == nil {
			continue
		}
		if err := r.processMessage(ctx, rdr); err != nil {
			log.Printf("imap receiver: process message error: %v", err)
		}
		// mark as seen
		toMark.AddNum(msg.SeqNum)
	}

	// store \Seen flag for processed messages
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	if toMark.Set != nil {
		if err := c.Store(&toMark, item, flags, nil); err != nil {
			// log but don't fail
			log.Printf("imap receiver: failed to mark seen: %v", err)
		}
	}

	return nil
}

// processMessage parses a single message body and creates DeliveryEvent records if bounce found.
func (r *IMAPReceiver) processMessage(ctx context.Context, body io.Reader) error {
	// parse with go-message mail reader
	mr, err := mail.CreateReader(body)
	if err != nil {
		// sometimes the body may include extra data; try buffered parse
		br := bufio.NewReader(body)
		mr2, err2 := mail.CreateReader(br)
		if err2 != nil {
			return err
		}
		mr = mr2
	}

	// Collect top-level headers
	h := mr.Header
	subject, _ := h.Subject()
	msgID, _ := h.MessageID()

	// read all body parts into a concatenated string for heuristics
	var fullBody strings.Builder
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
			fullBody.Write(b)
		case *mail.AttachmentHeader:
			// skip attachments
		default:
			b, _ := io.ReadAll(p.Body)
			fullBody.Write(b)
		}
	}

	bodyStr := fullBody.String()

	// Heuristics: check for known bounce headers or markers
	// 1) Final-Recipient header
	// 2) Action: failed
	// 3) X-Failed-Recipients header
	var bouncedRecipients []string
	var reason string

	// header checks (best-effort)
	if vals, err := h.AddressList("Final-Recipient"); err == nil && len(vals) > 0 {
		for _, v := range vals {
			bouncedRecipients = append(bouncedRecipients, v.Address)
		}
	}
	// X-Failed-Recipients custom header (may be present)
	if xvals, err := h.Text("X-Failed-Recipients"); err == nil && len(xvals) > 0 {
		for _, line := range xvals {
			sline := fmt.Sprint(line)
			for _, addr := range strings.Split(sline, ",") {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					bouncedRecipients = append(bouncedRecipients, addr)
				}
			}
		}
	}
	// fallback: search body for typical DSN markers
	lower := strings.ToLower(bodyStr)
	if strings.Contains(lower, "action: failed") || strings.Contains(lower, "status: 5.") {
		reason = "permanent failure (5xx)"
	}
	// attempt to extract recipient-looking tokens from body if none found
	if len(bouncedRecipients) == 0 {
		// look for "Final-Recipient: rfc822; user@example.com"
		for _, line := range strings.Split(bodyStr, "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(l), "final-recipient:") {
				parts := strings.Split(l, ";")
				if len(parts) > 1 {
					addr := strings.TrimSpace(parts[len(parts)-1])
					bouncedRecipients = append(bouncedRecipients, addr)
				}
			}
			if strings.HasPrefix(strings.ToLower(l), "x-failed-recipients:") {
				parts := strings.SplitN(l, ":", 2)
				if len(parts) == 2 {
					for _, addr := range strings.Split(parts[1], ",") {
						if a := strings.TrimSpace(addr); a != "" {
							bouncedRecipients = append(bouncedRecipients, a)
						}
					}
				}
			}
		}
	}

	// If still nothing that looks like a bounce, skip
	if len(bouncedRecipients) == 0 && reason == "" && !strings.Contains(lower, "undeliverable") && !strings.Contains(lower, "returned to sender") {
		// not identified as bounce
		return nil
	}

	now := time.Now().Unix()
	// Create event for each bounced recipient (best-effort)
	for _, rcpt := range bouncedRecipients {
		ev := &domain.DeliveryEvent{
			ID:         "", // repository may set ID
			DeliveryID: "", // unresolved here
			EventType:  domain.EventTypeBounced,
			EventData: map[string]interface{}{
				"recipient":      rcpt,
				"subject":        subject,
				"original_msgid": msgID,
				"raw_preview":    firstN(bodyStr, 1024),
			},
			CreatedAt: now,
		}
		if reason != "" {
			ev.EventData["reason"] = reason
		}
		// write synchronously
		if err := r.db.EventRepository().Create(ctx, ev); err != nil {
			log.Printf("imap receiver: failed to save event: %v", err)
		}
	}

	return nil
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
