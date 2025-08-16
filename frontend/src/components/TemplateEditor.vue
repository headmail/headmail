<!--
 Copyright 2025 JC-Lab
 SPDX-License-Identifier: AGPL-3.0-or-later
-->

<template>
  <div class="space-y-3">

    <!-- Subject input placed below (full-width under template name in parent view) -->
    <div class="mb-2">
      <input v-model="subjectInput" type="text" placeholder="Subject (optional)"
             class="w-full px-3 py-2 border rounded text-sm"/>
    </div>

    <div class="flex items-center gap-2 mb-2">
      <div class="flex items-center gap-2">
        <button
            type="button"
            :class="['px-3 py-1 rounded-md text-sm', activeTab === 'editor' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
            @click="activeTab = 'editor'">
          GrapesJS
        </button>
        <button
            type="button"
            :class="['px-3 py-1 rounded-md text-sm', activeTab === 'mjml' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
            @click="activeTab = 'mjml'">
          MJML
        </button>
        <button
            type="button"
            :class="['px-3 py-1 rounded-md text-sm', activeTab === 'preview' ? 'bg-blue-600 text-white' : 'bg-gray-100']"
            @click="activeTab = 'preview'">
          미리보기
        </button>
        <div class="ml-auto flex items-center gap-3">
          <div class="relative">
            <button type="button" @click="showSamples = !showSamples" class="px-2 py-1 rounded-md bg-gray-100 text-sm">
              Samples
            </button>
            <div v-if="showSamples" class="absolute right-0 mt-2 w-48 bg-white border rounded shadow z-10">
              <ul>
                <li
                    v-for="s in samples"
                    :key="s.id"
                    class="px-3 py-2 hover:bg-gray-100 cursor-pointer text-sm"
                    @click="loadSample(s)"
                >
                  {{ s.name }}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sample fields and preview button in a compact single row -->
    <div class="flex items-center gap-2 mb-2">
      <input v-model="sampleName" type="text" placeholder="Sample name" class="px-2 py-1 border rounded text-sm w-44"/>
      <input v-model="sampleEmail" type="text" placeholder="Sample email"
             class="px-2 py-1 border rounded text-sm w-64"/>
      <button
          type="button"
          @click="doServerPreview"
          :disabled="previewLoading"
          class="px-3 py-1 rounded-md text-sm bg-blue-600 text-white"
      >
        {{ previewLoading ? 'Rendering…' : 'Server Preview' }}
      </button>
      <div class="text-xs text-gray-500 ml-2 hidden sm:block">Tip: 드래그해서 레이아웃을 만드세요</div>
    </div>

    <div v-show="activeTab === 'editor'" class="border border-gray-200 rounded-lg overflow-hidden p-0"
         style="min-height:240px;">
      <div ref="editorContainer" :style="containerStyle"></div>
    </div>

    <div v-show="activeTab === 'mjml'">
      <textarea
          v-model="localMjml"
          rows="12"
          class="w-full px-4 py-3 border border-gray-300 rounded-xl font-mono text-sm"
          placeholder="MJML 형식의 이메일 본문을 입력하세요..."></textarea>
    </div>

    <div v-show="activeTab === 'preview'">
      <div class="w-full border border-gray-200 rounded-lg p-4 bg-white">
        <div v-if="serverPreviewSubject" class="mb-2 text-sm font-semibold">{{ serverPreviewSubject }}</div>
        <div v-if="previewError" class="text-sm text-red-600 mb-2">{{ previewError }}</div>
        <div v-if="!serverPreviewHtml && previewLoading" class="text-sm text-gray-600">Rendering preview...</div>

        <div v-if="serverPreviewHtml" v-html="serverPreviewHtml" class="min-h-[120px]"></div>
        <div v-else class="min-h-[120px]" v-html="localPreviewHtml"></div>

        <div v-if="serverPreviewText" class="mt-4 p-2 bg-gray-50 rounded text-sm font-mono whitespace-pre-wrap">
          {{ serverPreviewText }}
        </div>
      </div>
    </div>
    <div class="flex gap-2 mb-2">
      <!-- right side: compact controls removed from this row to keep single-row tabs -->
      <div class="text-xs text-gray-400 hidden sm:block">
        <div class="font-semibold text-sm mb-1">Available template variables</div>
        <ul class="list-inside list-disc">
          <li><span class="font-mono local-pre">{{formatVarName('.name')}}</span> — user name (string)</li>
          <li><span class="font-mono local-pre">{{formatVarName('.deliveryId')}}</span> — user email (string)</li>
          <li><span class="font-mono local-pre">{{formatVarName('.deliveryId')}}</span> — delivery identifier generated per-send (string)</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/* Minimal comments in English only per Dev Guide.
   Adds server preview UI and calls previewTemplate API (typed via generated types).
*/

import {ref, onMounted, onBeforeUnmount, watch, nextTick, computed} from 'vue';
import grapesjs from 'grapesjs';
import presetNewsletter from 'grapesjs-preset-newsletter';
import gjsMjml from 'grapesjs-mjml';
import 'grapesjs/dist/css/grapes.min.css';
import mjml2html from 'mjml-browser';
import {previewTemplate} from '../api';

const props = defineProps<{
  subject?: string;
  modelValueMjml?: string;
  modelValueHtml?: string;
  fullscreen?: boolean;
}>();

const formatVarName = (name: string) => {
  return `{{ ${name} }}`;
}

// compute container style so editor has an explicit height when modal is not fullscreen
const containerStyle = computed(() => {
  return {
    // use a fixed min-height when not fullscreen so GrapesJS canvas can calculate properly
    height: props.fullscreen ? '100%' : '420px',
    width: '100%',
    minWidth: '0',
  } as Record<string, string>;
});

const emits = defineEmits(['update:html', 'update:grapes', 'update:mjml', 'update:subject']);

const editorContainer = ref<HTMLElement | null>(null);
let editor: any = null;

const activeTab = ref<'editor' | 'mjml' | 'preview'>('editor');
const localMjml = ref(props.modelValueMjml || '');
const localPreviewHtml = computed(() => {
  if (!localMjml.value) {
    return '';
  }
  const res = mjml2html(localMjml.value, { keepComments: false });
  return (res && (res as any).html) || '';
})

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
        <mj-text align="center" color="#489BDA" font-size="25px" font-family="Arial, sans-serif" font-weight="bold" line-height="35px" padding-top="20px">You'll get a 15% discount <br />
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

const loadMjmlToEditor = () => {
  try {
    console.log('loadMjmlToEditor: ', localMjml.value)
    // const result = mjml2html(localMjml.value, {keepComments: false});
    // const compiledHtml = result && result.html ? result.html : '';
    editor.setComponents(localMjml.value);
  } catch (e) {
    console.error('Failed to compile MJML sample:', e);
  }
}

const loadSample = (s: { id: string; name: string; html?: string; mjml?: string }) => {
  showSamples.value = false;
  if (!editor) return;
  if (s.mjml) {
    localMjml.value = s.mjml;
    loadMjmlToEditor();
  }
};

// Server preview state
const sampleName = ref('John Doe');
const sampleEmail = ref('john@example.com');
const subjectInput = ref(props.subject || '');
const serverPreviewHtml = ref('');
const serverPreviewText = ref('');
const serverPreviewSubject = ref('');
const previewLoading = ref(false);
const previewError = ref('');

// Keep subject in sync with parent prop and notify parent on changes
watch(
  () => props.subject,
  (v) => {
    subjectInput.value = v || '';
  }
);

watch(subjectInput, (v) => {
  emits('update:subject', v);
});

/**
 * Request server-side rendering of the current template HTML + subject using sample data.
 * Uses previewTemplate from frontend API (typed against generated types).
 */
const doServerPreview = async () => {
  previewLoading.value = true;
  previewError.value = '';
  try {
    const req = {
      templateHtml: localPreviewHtml.value,
      templateText: '',
      subject: subjectInput.value,
      name: sampleName.value,
      email: sampleEmail.value,
    };
    const res = await previewTemplate(req as any);
    // response typed in generated types; guard access as defensive code
    serverPreviewHtml.value = (res && (res as any).html) || '';
    serverPreviewText.value = (res && (res as any).text) || '';
    serverPreviewSubject.value = (res && (res as any).subject) || '';
    activeTab.value = 'preview';
  } catch (e: any) {
    console.error('Preview failed', e);
    previewError.value = e?.message || String(e);
  } finally {
    previewLoading.value = false;
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
  if (props.modelValueMjml) {
    localMjml.value = props.modelValueMjml;
    loadMjmlToEditor();
  }

  // Update localHtml from editor and emit grapes project JSON
  const updateAll = () => {
    try {
      localMjml.value = editor.runCommand('mjml-code')
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

 // When switching to the editor tab, ensure GrapesJS refreshes to fit container and restore MJML-compiled HTML.
watch(
    () => activeTab.value,
    async (val) => {
      if (!editor) return;
      await nextTick();
      try {
        // If returning to editor and we have MJML source, compile it and restore into the editor.
        if (val === 'editor') {
          loadMjmlToEditor()
        }

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

/*
  Watch MJML source changes and update compiled HTML + editor when appropriate.
  - When user edits MJML (html tab), compile MJML -> HTML and:
    * update localHtml (compiled)
    * if editor is visible (editor tab) or user is editing the mjml tab, setComponents so preview stays in sync
*/
watch(localMjml, (v) => {
  try {
    if (!v || typeof mjml2html !== 'function') {
      return;
    }
    console.log(`localMjml watch: ${v}`)
    const res = mjml2html(v, { keepComments: false });
    const compiled = (res && (res as any).html) || '';
    // emit mjml and compiled html to parent
    emits('update:mjml', v);
    emits('update:html', compiled);
  } catch (e) {
    // ignore compilation errors
    console.log(`localMjml error: `, e);
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

.local-pre {
  display: inline-block;
  background-color: #f0f0f0;
  padding: 1px;
}
</style>
