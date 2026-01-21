function setPreviewContent(content) {
  const previewContent = document.getElementById('preview-content');
  const shadow = previewContent.shadowRoot || previewContent.attachShadow({ mode: 'open' });
  shadow.innerHTML = `<style>:host{ background-color: white; } :host { all: initial; display: block; background-color: white; }</style>${content}`;
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
