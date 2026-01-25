function setPreviewContent(content) {
  const previewContent = document.getElementById('preview-content');
  const shadow = previewContent.shadowRoot || previewContent.attachShadow({ mode: 'open' });
  shadow.innerHTML = `<link rel="preload" href="/static/fonts/inter/InterVariable.woff2" as="font" type="font/woff2" crossorigin>
    <link rel="preload" href="/static/fonts/inter/InterVariable-Italic.woff2" as="font" type="font/woff2" crossorigin>
    <link rel="stylesheet" href="/static/css/default.css">
    <style>
      :host {
        all: initial;
        display: block;
        background-color: #F4F1EA;
        color: #2D2A26;
        font-family: 'Inter', system-ui, -apple-system, sans-serif;
        font-size: 18px; 
        line-height: 1.8;
        padding: 2.5rem; 
        margin: 0;
        -webkit-font-smoothing: antialiased;
        font-optical-sizing: auto;
      }
    </style>${content}`;
}

function updateSlug(value) {
  document.getElementById('slug').value = toKebabCase(value);
}

function toKebabCase(str) {
  return str
    .toLowerCase()
    .trim()
    .replace(/[çÇ]/g, 'c')
    .replace(/[ğĞ]/g, 'g')
    .replace(/[ıİ]/g, 'i')
    .replace(/[öÖ]/g, 'o')
    .replace(/[şŞ]/g, 's')
    .replace(/[üÜ]/g, 'u')
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}
