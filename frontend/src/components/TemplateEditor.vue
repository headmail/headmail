<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2 mb-2">
      <button
        type="button"
        :class="['px-3 py-1 rounded-md text-sm', activeTab === 'editor' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
        @click="activeTab = 'editor'">
        GrapesJS
      </button>
      <button
        type="button"
        :class="['px-3 py-1 rounded-md text-sm', activeTab === 'html' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
        @click="activeTab = 'html'">
        HTML
      </button>
      <button
        type="button"
        :class="['px-3 py-1 rounded-md text-sm', activeTab === 'preview' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
        @click="activeTab = 'preview'">
        미리보기
      </button>
      <div class="ml-auto flex items-center space-x-3">
        <div class="relative">
          <button type="button" @click="showSamples = !showSamples" class="px-3 py-1 rounded-md bg-gray-100 text-sm">
            Samples
          </button>
          <div v-if="showSamples" class="absolute right-0 mt-2 w-56 bg-white border rounded shadow z-10">
            <ul>
              <li
                v-for="s in samples"
                :key="s.id"
                class="px-3 py-2 hover:bg-gray-100 cursor-pointer"
                @click="loadSample(s)"
              >
                {{ s.name }}
              </li>
            </ul>
          </div>
        </div>
        <div class="text-xs text-gray-500">Tip: 드래그해서 레이아웃을 만드세요</div>
      </div>
    </div>

    <div v-show="activeTab === 'editor'" class="border border-gray-200 rounded-lg overflow-hidden p-0" style="min-height:240px;">
      <div ref="editorContainer" :style="containerStyle"></div>
    </div>

    <div v-show="activeTab === 'html'">
      <textarea
        v-model="localHtml"
        rows="12"
        class="w-full px-4 py-3 border border-gray-300 rounded-xl font-mono text-sm"
        placeholder="HTML 형식의 이메일 본문을 입력하세요..."></textarea>
    </div>

    <div v-show="activeTab === 'preview'">
      <div class="w-full border border-gray-200 rounded-lg p-4 bg-white" v-html="localHtml"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue';
import grapesjs from 'grapesjs';
import presetNewsletter from 'grapesjs-preset-newsletter';
import gjsMjml from 'grapesjs-mjml';
import 'grapesjs/dist/css/grapes.min.css';
import mjml2html from 'mjml-browser';

const props = defineProps<{
  modelValueHtml?: string;
  modelValueGrapes?: string | Record<string, any>;
  modelValueMjml?: string;
  fullscreen?: boolean;
}>();

 // compute container style so editor has an explicit height when modal is not fullscreen
 const containerStyle = computed(() => {
   return {
     // use a fixed min-height when not fullscreen so GrapesJS canvas can calculate properly
     height: props.fullscreen ? '100%' : '420px',
     width: '100%',
     minWidth: '0',
   } as Record<string, string>;
 });

const emits = defineEmits(['update:html', 'update:grapes', 'update:mjml']);

const editorContainer = ref<HTMLElement | null>(null);
let editor: any = null;

const activeTab = ref<'editor' | 'html' | 'preview'>('editor');
const localHtml = ref(props.modelValueHtml || '');
const initialGrapes = props.modelValueGrapes || null;

 // Samples dropdown state and sample definitions (simple HTML snippets).
 const showSamples = ref(false);
 const samples: Array<{ id: string; name: string; html?: string; mjml?: string }> = [
  {
    id: 'newsletter',
    name: 'MJML - Newsletter (Hero + CTA)',
    mjml: `
      <mjml>
        <mj-body background-color="#f3f4f6">
          <mj-section>
            <mj-column>
              <mj-text font-size="20px" font-weight="bold">Monthly Newsletter</mj-text>
              <mj-text color="#374151">Short intro paragraph — summarize updates and highlights.</mj-text>
              <mj-button background-color="#3b82f6" color="#ffffff">Read More</mj-button>
            </mj-column>
          </mj-section>
          <mj-section>
            <mj-column>
              <mj-text font-size="12px" color="#9ca3af">You are receiving this email because you signed up for updates.</mj-text>
            </mj-column>
          </mj-section>
        </mj-body>
      </mjml>
    `.trim()
  },
  {
    id: 'promo',
    name: 'mjml.io - happy-new-year',
    mjml: `
      <mjml version="3.3.3">
  <mj-body background-color="#F4F4F4" color="#55575d" font-family="Arial, sans-serif">
    <mj-section background-color="#C1272D" background-repeat="repeat" padding="20px 0" text-align="center" vertical-align="top">
      <mj-column>
        <mj-image align="center" padding="10px 25px" src="http://gkq4.mjt.lu/img/gkq4/b/18rxz/1h3k4.png" width="128px"></mj-image>
      </mj-column>
    </mj-section>
    <mj-section background-color="#ffffff" background-repeat="repeat" padding="20px 0" text-align="center" vertical-align="top">
      <mj-column>
        <mj-image align="center" padding="10px 25px" src="http://gkq4.mjt.lu/img/gkq4/b/18rxz/1h3s5.gif" width="600px"></mj-image>
        <mj-image align="center" alt="Happy New Year!" container-background-color="#ffffff" padding="10px 25px" src="http://gkq4.mjt.lu/img/gkq4/b/18rxz/1hlvp.png" width="399px"></mj-image>
      </mj-column>
    </mj-section>
    <mj-section background-color="#ffffff" background-repeat="repeat" background-size="auto" padding="20px 0px 20px 0px" text-align="center" vertical-align="top">
      <mj-column>
        <mj-text align="center" color="#55575d" font-family="Arial, sans-serif" font-size="14px" line-height="28px" padding="0px 25px 0px 25px">New dreams, new hopes, new experiences and new joys, we wish you all the best for this New Year to come in 2018!</mj-text>
        <mj-image align="center" alt="Best wishes from all the Clothes Team!" padding="10px 25px" src="http://gkq4.mjt.lu/img/gkq4/b/18rxz/1hlv8.png" width="142px"></mj-image>
      </mj-column>
    </mj-section>
    <mj-section background-color="#C1272D" background-repeat="repeat" padding="20px 0" text-align="center" vertical-align="top">
      <mj-column>
        <mj-text align="center" color="#ffffff" font-family="Arial, sans-serif" font-size="13px" line-height="22px" padding="10px 25px">Simply created&nbsp;on&nbsp;<a style="color:#ffffff" href="http://www.mailjet.com"><b>Mailjet Passport</b></a></mj-text>
      </mj-column>
    </mj-section>
    <mj-section background-repeat="repeat" background-size="auto" padding="20px 0px 20px 0px" text-align="center" vertical-align="top">
      <mj-column>
        <mj-text align="center" color="#55575d" font-family="Arial, sans-serif" font-size="11px" line-height="22px" padding="0px 20px">[[DELIVERY_INFO]]</mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
    `.trim()
  },
  {
    id: 'simple',
    name: 'mjml.io - referral',
    mjml: `
<mjml>
  <mj-head>
    <mj-attributes>
      <mj-all padding="0px"></mj-all>
      <mj-text font-family="Ubuntu, Helvetica, Arial, sans-serif" padding="0 25px" font-size="13px"></mj-text>
      <mj-section background-color="#ffffff"></mj-section>
      <mj-class name="preheader" color="#000000" font-size="11px"></mj-class>
    </mj-attributes>
    <mj-style inline="inline">a { text-decoration: none!important; color: inherit!important; }</mj-style>
  </mj-head>
  <mj-body background-color="#bedae6">
    <mj-section>
      <mj-column width="100%">
        <mj-image src="http://go.mailjet.com/tplimg/mtrq/b/ox8s/mg1q9.png" alt="header image" padding="0px"></mj-image>
      </mj-column>
    </mj-section>
    <mj-section padding-bottom="20px" padding-top="10px">
      <mj-column>
        <mj-text align="center" padding="10px 25px" font-size="20px" color="#512d0b"><strong>Hey {{FirstName}}!</strong></mj-text>
        <mj-text align="center" font-size="18px" font-family="Arial">Are you enjoying our weekly newsletter?<br /> Then why not share it with your friends?</mj-text>
        <mj-text align="center" color="#489BDA" font-size="25px" font-family="Arial, sans-serif" font-weight="bold" line-height="35px" padding-top="20px">You&apos;ll get a 15% discount <br />
          <span style="font-size:18px">on your next order when a friend uses the code {{ReferalCode}}!</span>
        </mj-text>
        <mj-button background-color="#8bb420" color="#FFFFFF" href="https://mjml.io" font-family="Arial, sans-serif" padding="20px 0 0 0" font-weight="bold" font-size="16px">Refer a friend now</mj-button>
        <mj-text align="center" color="#000000" font-size="14px" font-family="Arial, sans-serif" padding-top="40px">Best, <br /> The {{CompanyName}} Team
          <p></p>
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
    `.trim()
  },
  {
    id: 'realistic',
    name: 'MJML - Realistic Newsletter (Header, 2-column, footer)',
    mjml: `
      <mjml>
        <mj-body background-color="#f5f7fb">
          <mj-section background-color="#ffffff" padding="20px">
            <mj-column width="100%">
              <mj-image src="https://via.placeholder.com/150x40?text=Logo" alt="Logo" width="150px"></mj-image>
            </mj-column>
          </mj-section>

          <mj-section background-color="#ffffff" padding="20px">
            <mj-column width="50%">
              <mj-text font-size="18px" font-weight="bold">Product Update</mj-text>
              <mj-text color="#374151">We launched new features this month. Here's a summary.</mj-text>
              <mj-button background-color="#2563eb" color="#ffffff">View Details</mj-button>
            </mj-column>
            <mj-column width="50%">
              <mj-image src="https://via.placeholder.com/260x140" alt="Feature" />
            </mj-column>
          </mj-section>

          <mj-section padding="20px">
            <mj-column>
              <mj-text font-size="12px" color="#9aa0a6">If you no longer wish to receive these emails, you can <a href="#">unsubscribe</a>.</mj-text>
            </mj-column>
          </mj-section>
        </mj-body>
      </mjml>
    `.trim()
  }
];

const loadSample = (s: { id: string; name: string; html?: string; mjml?: string }) => {
  showSamples.value = false;
  if (!editor) return;
  try {
    if (s.html) {
      editor.setComponents(s.html);
      // small style to make buttons look nicer
      try {
        editor.setStyle('body { font-family: Arial, sans-serif; } a { text-decoration:none; }');
      } catch {}
      // emit immediately
      const html = editor.getHtml();
      localHtml.value = html;
      emits('update:html', html);
      emits('update:grapes', JSON.stringify({
        html: editor.getHtml(),
        css: editor.getCss ? editor.getCss() : '',
        components: editor.getComponents ? editor.getComponents() : '',
        style: editor.getStyle ? editor.getStyle() : ''
      }));
    }
    if (s.mjml) {
      // compile MJML to HTML then load into editor
      try {
        const result = mjml2html(s.mjml, { keepComments: false });
        const compiledHtml = result && result.html ? result.html : '';
        editor.setComponents(compiledHtml);
        // set MJML value to parent via emit so it's stored
        emits('update:mjml', s.mjml);
        // emit compiled HTML and grapes project snapshot
        localHtml.value = compiledHtml;
        emits('update:html', compiledHtml);
        emits('update:grapes', JSON.stringify({
          html: editor.getHtml(),
          css: editor.getCss ? editor.getCss() : '',
          components: editor.getComponents ? editor.getComponents() : '',
          style: editor.getStyle ? editor.getStyle() : ''
        }));
      } catch (e) {
        console.error('Failed to compile MJML sample:', e);
      }
    }
  } catch (e) {
    console.error('Failed to load sample:', e);
  }
};

// Initialize GrapesJS editor
onMounted(async () => {
  await nextTick();
  editor = grapesjs.init({
    container: editorContainer.value as HTMLElement,
    fromElement: false,
    height: '100%',
    width: 'auto',
    storageManager: {
      autoload: false,
      autosave: false,
    },
    plugins: [presetNewsletter, gjsMjml],
    pluginsOpts: {
      [presetNewsletter as any]: {},
      [gjsMjml as any]: {},
    },
  });

  // Load initial content: prefer grapes project if provided, otherwise load HTML.
  // If nothing provided, load a small default sample so users can start quickly.
  try {
    if (initialGrapes) {
      let data: any = initialGrapes;
      if (typeof initialGrapes === 'string' && initialGrapes.trim() !== '') {
        data = JSON.parse(initialGrapes);
      }
      if (data) {
        if (data.components) {
          editor.setComponents(data.components);
        } else if (data.html) {
          editor.setComponents(data.html);
        } else if (typeof data === 'string') {
          editor.setComponents(data);
        }
        if (data.style) {
          editor.setStyle(data.style);
        }
        if (data.css) {
          editor.setCss ? editor.setCss(data.css) : null;
        }
      }
    } else if (props.modelValueHtml) {
      editor.setComponents(props.modelValueHtml);
    } else {
      // No initial content provided — load a helpful sample template
      const sampleHtml = `
        <table width="100%" cellpadding="0" cellspacing="0" role="presentation">
          <tr>
            <td align="center" style="padding:32px 16px;background:#f8fafc;">
              <table width="600" cellpadding="0" cellspacing="0" role="presentation" style="max-width:600px;">
                <tr>
                  <td style="padding:24px 32px;background:#ffffff;border-radius:8px;text-align:left;font-family:Arial, sans-serif;">
                    <h1 style="margin:0 0 12px;font-size:24px;color:#111827;">Welcome to Headmail</h1>
                    <p style="margin:0 0 18px;color:#374151;">This is a sample newsletter template. Use the GrapesJS blocks to customize layout, images and buttons.</p>
                    <p style="margin:0;"><a href=\"#\" class=\"btn\" style=\"display:inline-block;padding:10px 18px;background:#3b82f6;color:#ffffff;border-radius:6px;text-decoration:none;\">Get Started</a></p>
                  </td>
                </tr>
                <tr>
                  <td style="padding:12px 32px;text-align:center;font-size:12px;color:#9ca3af;">You are receiving this email because you signed up for updates.</td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
      `.trim();

      // Set sample HTML and a minimal style so it looks decent in preview/editor.
      editor.setComponents(sampleHtml);
      try {
        editor.setStyle('body { font-family: Arial, sans-serif; } .btn { display:inline-block; padding:10px 18px; background:#3b82f6; color:#fff; border-radius:6px; text-decoration:none;}');
      } catch {}
    }
  } catch (e) {
    // fallback: load html
    if (props.modelValueHtml) {
      editor.setComponents(props.modelValueHtml);
    }
    console.error('Failed to load grapes project:', e);
  }

  // Update localHtml from editor and emit grapes project JSON
  const updateAll = () => {
    try {
      const html = editor.getHtml();
      localHtml.value = html;
      emits('update:html', html);

      // build grapes project JSON
      let grapesData: any = {};
      if (editor.getProjectData) {
        grapesData = editor.getProjectData();
      } else {
        grapesData = {
          html: editor.getHtml(),
          css: editor.getCss ? editor.getCss() : '',
          components: editor.getComponents ? editor.getComponents() : '',
          style: editor.getStyle ? editor.getStyle() : '',
        };
      }
      emits('update:grapes', JSON.stringify(grapesData));
    } catch (err) {
      console.error('Error updating editor content:', err);
    }
  };

  editor.on && editor.on('update', updateAll);
  // also listen to component:add/remove/change for better coverage
  editor.on && editor.on('component:add', updateAll);
  editor.on && editor.on('component:remove', updateAll);
  editor.on && editor.on('style:change', updateAll);

  // initial emit
  updateAll();
});


// When fullscreen prop changes, re-render / refresh editor so canvas resizes correctly.
watch(
  () => props.fullscreen,
  async () => {
    if (!editor) return;
    await nextTick();
    try {
      editor.render && editor.render();
      // ensure iframe frame element uses full width
      if (editor.Canvas && editor.Canvas.getFrameEl) {
        const frame = editor.Canvas.getFrameEl();
        if (frame && (frame as HTMLElement).style) {
          (frame as HTMLElement).style.width = '100%';
          (frame as HTMLElement).style.minWidth = '0';
        }
      }
      // dispatch resize to help internal layouts recalc
      window.dispatchEvent(new Event('resize'));
    } catch (e) {
      // ignore
    }
  }
);

// When switching to the editor tab, ensure GrapesJS refreshes to fit container.
watch(
  () => activeTab.value,
  async (val) => {
    if (val !== 'editor' || !editor) return;
    await nextTick();
    try {
      editor.render && editor.render();
      if (editor.Canvas && editor.Canvas.getFrameEl) {
        const frame = editor.Canvas.getFrameEl();
        if (frame && (frame as HTMLElement).style) {
          (frame as HTMLElement).style.width = '100%';
          (frame as HTMLElement).style.minWidth = '0';
        }
      }
      window.dispatchEvent(new Event('resize'));
      // small delay refresh for tricky layouts
      setTimeout(() => {
        try {
          editor.render && editor.render();
        } catch {}
      }, 50);
    } catch (e) {
      // ignore
    }
  }
);

onBeforeUnmount(() => {
  try {
    if (editor && editor.destroy) {
      editor.destroy();
      editor = null;
    }
  } catch (e) {
    // ignore
  }
});

// Watch the external html prop and update local editor if needed
watch(
  () => props.modelValueHtml,
  (v) => {
    if (v !== undefined && v !== localHtml.value && editor) {
      localHtml.value = v as string;
      try {
        editor.setComponents(v);
      } catch (e) {
        // ignore
      }
    }
  }
);

// Watch localHtml changes (from textarea) and push to editor when in HTML tab
watch(localHtml, (v) => {
  if (editor && activeTab.value === 'html') {
    try {
      editor.setComponents(v);
    } catch (e) {
      // ignore
    }
    // emit html change and grapes snapshot
    emits('update:html', v);
    const grapesData = {
      html: v,
      css: editor && editor.getCss ? editor.getCss() : '',
      components: editor && editor.getComponents ? editor.getComponents() : '',
      style: editor && editor.getStyle ? editor.getStyle() : '',
    };
    emits('update:grapes', JSON.stringify(grapesData));
  }
});
</script>

<style scoped>
.gjs-blocks-c {
  max-height: 420px;
  overflow: auto;
}

/* Ensure GrapesJS editor and iframe are responsive inside modal/container */
::v-deep .gjs {
  width: 100% !important;
  box-sizing: border-box;
}
::v-deep .gjs-editor {
  width: 100% !important;
  min-width: 0;
}
::v-deep .gjs-frame {
  width: 100% !important;
  min-width: 0;
}
::v-deep .gjs-cv {
  max-width: none !important;
}
</style>
