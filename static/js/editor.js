import { Editor, Node } from 'https://esm.sh/@tiptap/core@2'
import StarterKit from 'https://esm.sh/@tiptap/starter-kit@2'
import TextAlign from 'https://esm.sh/@tiptap/extension-text-align@2'
import Link from 'https://esm.sh/@tiptap/extension-link@2'
import Image from 'https://esm.sh/@tiptap/extension-image@2'

const Figure = Node.create({
  name: 'figure',
  group: 'block',
  content: 'figcaption',
  draggable: true,
  isolating: true,

  addAttributes() {
    return {
      src: { default: null },
      alt: { default: null },
    }
  },

  parseHTML() {
    return [{
      tag: 'figure',
      contentElement: 'figcaption',
      getAttrs: dom => {
        const img = dom.querySelector('img')
        return {
          src: img?.getAttribute('src'),
          alt: img?.getAttribute('alt'),
        }
      },
    }]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      'figure',
      ['img', { src: HTMLAttributes.src, alt: HTMLAttributes.alt, 'data-drag-handle': '' }],
      ['figcaption', 0],
    ]
  },
})

const Figcaption = Node.create({
  name: 'figcaption',
  group: 'content',
  content: 'inline*',
  parseHTML() { return [{ tag: 'figcaption' }] },
  renderHTML() { return ['figcaption', 0] },
  selectable: false,
})


const initialContent = document.getElementById('editor-initial-content').value
const contentInput = document.getElementById('content-input')
const toolbar = document.getElementById('editor-toolbar')
const fileInput = document.getElementById('image-file-input')

const editor = new Editor({
  element: document.getElementById('editor'),
  extensions: [
    StarterKit,
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    Link.configure({ openOnClick: false }),
    Image,
    Figure,
    Figcaption,
  ],
  content: initialContent || '',
  onUpdate() {
    contentInput.value = editor.getHTML()
  },
})

// Expose для E2E-тестов (CHK-01: setContent без keyboard interactions)
window._tiptapEditor = editor

// Sync content on submit (на случай если onUpdate не успел)
document.querySelector('form.article-form').addEventListener('submit', () => {
  contentInput.value = editor.getHTML()
})

// Toolbar button actions
toolbar.addEventListener('click', (e) => {
  const btn = e.target.closest('button[data-action]')
  if (!btn) return
  const action = btn.dataset.action

  switch (action) {
    case 'bold':        editor.chain().focus().toggleBold().run(); break
    case 'italic':      editor.chain().focus().toggleItalic().run(); break
    case 'strike':      editor.chain().focus().toggleStrike().run(); break
    case 'h1':          editor.chain().focus().toggleHeading({ level: 1 }).run(); break
    case 'h2':          editor.chain().focus().toggleHeading({ level: 2 }).run(); break
    case 'h3':          editor.chain().focus().toggleHeading({ level: 3 }).run(); break
    case 'link': {
      const url = window.prompt('URL ссылки:')
      if (url) editor.chain().focus().setLink({ href: url }).run()
      break
    }
    case 'align-left':   editor.chain().focus().setTextAlign('left').run(); break
    case 'align-center': editor.chain().focus().setTextAlign('center').run(); break
    case 'align-right':  editor.chain().focus().setTextAlign('right').run(); break
    case 'image-upload': fileInput.click(); break
  }
})

// Update toolbar active states on selection change
editor.on('selectionUpdate', updateToolbarState)
editor.on('update', updateToolbarState)

function updateToolbarState() {
  toolbar.querySelectorAll('button[data-action]').forEach(btn => {
    const a = btn.dataset.action
    let active = false
    if (a === 'bold')         active = editor.isActive('bold')
    else if (a === 'italic')  active = editor.isActive('italic')
    else if (a === 'strike')  active = editor.isActive('strike')
    else if (a === 'h1')      active = editor.isActive('heading', { level: 1 })
    else if (a === 'h2')      active = editor.isActive('heading', { level: 2 })
    else if (a === 'h3')      active = editor.isActive('heading', { level: 3 })
    else if (a === 'align-left')   active = editor.isActive({ textAlign: 'left' })
    else if (a === 'align-center') active = editor.isActive({ textAlign: 'center' })
    else if (a === 'align-right')  active = editor.isActive({ textAlign: 'right' })
    btn.classList.toggle('is-active', active)
  })
}

// Image upload: POST to endpoint, parse JSON response, insert figure node.
fileInput.addEventListener('change', async () => {
  const file = fileInput.files[0]
  if (!file) return

  const articleId = window.location.pathname.match(/\/admin\/articles\/(\d+)/)?.[1]
  if (!articleId) return

  const csrf = document.querySelector('input[name="_csrf"]').value
  const formData = new FormData()
  formData.append('image', file)
  formData.append('_csrf', csrf)

  try {
    const res = await fetch(`/admin/articles/${articleId}/images?from=tiptap`, {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrf },
      body: formData,
    })
    if (!res.ok) { alert('Ошибка загрузки изображения'); return }

    const data = await res.json()
    if (!data.url) { alert('Не удалось получить URL изображения'); return }

    const captionText = window.prompt('Подпись к изображению (необязательно):') || ''

    editor.chain().focus().insertContent({
      type: 'figure',
      attrs: { src: data.url, alt: captionText },
      content: [{
        type: 'figcaption',
        content: captionText ? [{ type: 'text', text: captionText }] : [],
      }],
    }).run()

  } catch {
    alert('Ошибка загрузки изображения')
  } finally {
    fileInput.value = ''
  }
})
