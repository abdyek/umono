function setPreviewContent(content) {
  const previewContent = document.getElementById('preview-content');
  const shadow = previewContent.shadowRoot || previewContent.attachShadow({ mode: 'open' });
  shadow.innerHTML = `<style>:host{ background-color: white; } :host { all: initial; display: block; background-color: white; }</style>${content}`;
}
