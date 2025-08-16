// Copyright 2025 JC-Lab
// SPDX-License-Identifier: AGPL-3.0-or-later

package imap

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-message/textproto"
	"github.com/headmail/headmail/pkg/config"
	"github.com/headmail/headmail/pkg/mailer"
	"github.com/headmail/headmail/pkg/receiver"
	"github.com/pkg/errors"
)

// Receiver polls an IMAP mailbox and records bounce events.
type Receiver struct {
	cfg        config.IMAPConfig
	mailbox    string
	pollPeriod time.Duration

	cancelCtx context.CancelFunc
	doneCh    chan struct{}
}

var doubleMessageIdPattern = regexp.MustCompile("^<<(.+)>>$")

// NewReceiver creates a new Receiver.
// cfg is the SMTP.Receive section from application config.
func NewReceiver(cfg *config.IMAPConfig) *Receiver {
	period := time.Second * 5
	return &Receiver{
		cfg:        *cfg,
		mailbox:    cfg.Mailbox,
		pollPeriod: period,
		doneCh:     make(chan struct{}),
	}
}

// Start begins polling the IMAP mailbox in background.
func (r *Receiver) Start(ctx context.Context) (chan *receiver.Event, error) {
	output := make(chan *receiver.Event, 1)
	ctx, cancel := context.WithCancel(ctx)
	r.cancelCtx = cancel

	mbox := r.mailbox
	if mbox == "" {
		mbox = "INBOX"
	}

	go func() {
		defer close(output)
		defer close(r.doneCh)

		ticker := time.NewTicker(r.pollPeriod)
		defer ticker.Stop()

		for ctx.Err() == nil {
			if err := r.loopOnce(ctx, mbox, ticker, output); err != nil {
				log.Printf("mail read failed: %+v", err)

				select {
				case <-ctx.Done():
				case <-ticker.C:
				}
			}
		}
	}()

	return output, nil
}

func (r *Receiver) Stop(ctx context.Context) error {
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

func (r *Receiver) loopOnce(ctx context.Context, mbox string, ticker *time.Ticker, output chan *receiver.Event) error {
	c, err := r.dial()
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}
	defer c.Logout()
	defer c.Close()

	if _, err := c.Select(mbox, false); err != nil {
		return errors.Wrap(err, "select mailbox failed")
	}

	if err := r.pollOnce(ctx, c, output); err != nil {
		log.Printf("imap receiver: poll error: %v", err)
	}

	stop := make(chan struct{})
	idleDone := make(chan error)
	go func() {
		defer close(idleDone)
		idleDone <- c.Idle(stop, nil)
	}()

	isIdleDone := false
	select {
	case <-ctx.Done():
	case <-ticker.C:
	case <-idleDone:
		isIdleDone = true
	}
	close(stop)
	if !isIdleDone {
		<-idleDone
	}

	return nil
}

// dial connects to IMAP server (TLS if configured).
func (r *Receiver) dial() (*imapclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", r.cfg.Host, r.cfg.Port)
	if r.cfg.TLS {
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
	supportStartTls, _ := c.SupportStartTLS()
	if supportStartTls {
		if err = c.StartTLS(nil); err != nil {
			log.Printf("StartTLS failed: %+v", err)
		}
	}

	// login
	if err := c.Login(r.cfg.Username, r.cfg.Password); err != nil {
		_ = c.Close()
		log.Printf("imap receiver: login error: %v", err)
		return nil, err
	}

	return c, nil
}

// pollOnce connects to IMAP, fetches UNSEEN messages and processes them.
func (r *Receiver) pollOnce(ctx context.Context, c *imapclient.Client, eventCh chan *receiver.Event) error {
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
		if err := r.processMessage(ctx, rdr, eventCh); err != nil {
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
func (r *Receiver) processMessage(ctx context.Context, body io.Reader, eventCh chan *receiver.Event) error {
	event := &receiver.Event{}

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
	event.Subject, _ = h.Subject()
	event.MessageID, _ = h.MessageID()
	if event.MessageID == "" {
		if matches := doubleMessageIdPattern.FindStringSubmatch(h.Get("Message-Id")); len(matches) > 0 {
			event.MessageID = matches[1]
		}
	}

	var deliveryStatusBody mail.Header
	var sentEntity *message.Entity

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch p.Header.(type) {
		case *mail.AttachmentHeader:
			contentType := p.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "message/delivery-status") {
				r := bufio.NewReader(p.Body)
				for {
					header, err := textproto.ReadHeader(r)
					if err != nil || header.Len() == 0 {
						break
					}
					for k, vs := range header.Map() {
						for _, v := range vs {
							deliveryStatusBody.Add(k, v)
						}
					}
				}
			} else if strings.HasPrefix(contentType, "message/rfc822") {
				sentEntity, err = message.Read(p.Body)
				if err != nil {
					return err
				}
			} else {
				_, _ = io.Copy(io.Discard, p.Body)
			}
		default:
			_, _ = io.Copy(io.Discard, p.Body)
		}
	}

	if deliveryStatusBody.Len() == 0 || sentEntity == nil {
		return nil
	}

	// Heuristics: check for known bounce headers or markers
	// 1) Final-Recipient header
	// 2) Action: failed
	// 3) X-Failed-Recipients header
	// event.Reason

	event.DeliveryID = sentEntity.Header.Get(mailer.HeadmailDeliveryHeaderName)

	// header checks (best-effort)
	if val, err := deliveryStatusBody.Text("Final-Recipient"); err == nil && len(val) > 0 {
		parsed, err := parseAddress(val)
		if err == nil {
			event.BouncedRecipients = append(event.BouncedRecipients, parsed)
		}
	}
	// X-Failed-Recipients custom header (may be present)
	if xvals, err := deliveryStatusBody.Text("X-Failed-Recipients"); err == nil && len(xvals) > 0 {
		for _, line := range xvals {
			sline := fmt.Sprint(line)
			for _, addr := range strings.Split(sline, ",") {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					event.BouncedRecipients = append(event.BouncedRecipients, addr)
				}
			}
		}
	}
	if val, err := deliveryStatusBody.Text("Status"); err == nil && strings.HasPrefix(val, "5.") {
		event.Reason = "permanent failure (" + val + ")"
	} else if val, err := deliveryStatusBody.Text("Action"); err == nil && val == "failed" {
		event.Reason = "permanent failure"
	}

	// If still nothing that looks like a bounce, skip
	if len(event.BouncedRecipients) == 0 && event.Reason == "" {
		// not identified as bounce
		return nil
	}
	if event.DeliveryID == "" {
		log.Printf("unknown delivery id: %+v", event)
		return nil
	}

	eventCh <- event

	return nil
}
