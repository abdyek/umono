function setPreviewContent(content) {
  const previewContent = document.getElementById('preview-content');
  const shadow = previewContent.shadowRoot || previewContent.attachShadow({ mode: 'open' });
  const previewHeight = `${previewContent.clientHeight}px`;
  shadow.innerHTML = `<link rel="preload" href="/static/fonts/inter/InterVariable.woff2" as="font" type="font/woff2" crossorigin>
    <link rel="preload" href="/static/fonts/inter/InterVariable-Italic.woff2" as="font" type="font/woff2" crossorigin>
    <link rel="stylesheet" href="/static/css/default.css">
    <style>
      :host {
        all: initial;
        display: block;
        --preview-height: ${previewHeight};
      }

      .preview-body {
        background-color: #F4F1EA;
        color: #2D2A26;
        font-family: 'Inter', system-ui, -apple-system, sans-serif;
        font-size: 18px; 
        line-height: 1.8;
        padding: 2.5rem; 
        margin: 0;
        min-height: var(--preview-height);
        box-sizing: border-box;
        -webkit-font-smoothing: antialiased;
        font-optical-sizing: auto;
      }

      .preview-body > compono-web-grid:nth-of-type(1) {
        min-height: calc(var(--preview-height) - 5rem);
      }
    </style><div class="preview-body">${content}</div>`;
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

function findSelfOrDescendant(root, selector) {
  if (!root) {
    return null;
  }

  if (root.matches && root.matches(selector)) {
    return root;
  }

  if (!root.querySelector) {
    return null;
  }

  return root.querySelector(selector);
}

function initComponentNameGuidance(root = document) {
  const input = findSelfOrDescendant(root, '[data-component-name-input]');
  const helper = findSelfOrDescendant(root, '[data-component-name-helper]');
  const helperDot = findSelfOrDescendant(root, '[data-component-name-helper-dot]');
  const error = findSelfOrDescendant(root, '[data-component-name-error]');

  if (!input || !helper || !helperDot || !error || input.dataset.componentNameBound === 'true') {
    return;
  }

  const componentNamePattern = /^[A-Z0-9]+(?:_[A-Z0-9]+)*$/;
  const mode = input.dataset.componentNameMode;
  const hasServerError = input.dataset.componentNameHasServerError === 'true';
  let hasInteracted = false;

  const updateState = () => {
    const value = input.value.trim();
    const isNeutral = value === '';
    const isValid = componentNamePattern.test(value);
    const shouldShow = mode === 'create' || (hasInteracted && !isNeutral && !isValid);

    input.classList.remove('border-neutral-700', 'border-emerald-500/40', 'border-amber-500/40', 'border-red-500/50');
    error.classList.remove('hidden', 'flex');
    helper.classList.remove('hidden', 'flex', 'items-center', 'gap-1.5', 'text-neutral-500', 'text-emerald-400', 'text-red-300', 'border-transparent', 'border-neutral-800', 'border-emerald-500/20', 'border-red-500/30', 'bg-transparent', 'bg-neutral-900/40', 'bg-emerald-500/10', 'bg-red-500/10');
    helperDot.classList.remove('bg-neutral-600', 'bg-emerald-400', 'bg-red-400');

    if (!hasInteracted || isNeutral) {
      if (hasServerError) {
        input.classList.add('border-red-500/50');
        error.classList.add('flex');
        helper.classList.add('hidden', 'border-transparent', 'bg-transparent');
      } else {
        input.classList.add('border-neutral-700');
        error.classList.add('hidden');
        if (shouldShow) {
          helper.classList.add('flex', 'items-center', 'gap-1.5', 'text-neutral-500', 'border-neutral-800', 'bg-neutral-900/40');
          helperDot.classList.add('bg-neutral-600');
        } else {
          helper.classList.add('hidden', 'border-transparent', 'bg-transparent');
        }
      }
      input.removeAttribute('aria-invalid');
      return;
    }

    error.classList.add('hidden');

    if (isValid) {
      input.classList.add('border-emerald-500/40');
      if (shouldShow) {
        helper.classList.add('flex', 'items-center', 'gap-1.5', 'text-emerald-400', 'border-emerald-500/20', 'bg-emerald-500/10');
        helperDot.classList.add('bg-emerald-400');
      } else {
        helper.classList.add('hidden', 'border-transparent', 'bg-transparent');
      }
      input.setAttribute('aria-invalid', 'false');
      return;
    }

    input.classList.add('border-amber-500/40');
    helper.classList.add('flex', 'items-center', 'gap-1.5', 'text-red-300', 'border-red-500/30', 'bg-red-500/10');
    helperDot.classList.add('bg-red-400');
    input.setAttribute('aria-invalid', 'true');
  };

  input.dataset.componentNameBound = 'true';
  input.addEventListener('input', () => {
    hasInteracted = true;
    updateState();
  });
  updateState();
}

document.addEventListener('DOMContentLoaded', () => {
  initComponentNameGuidance();
});

document.addEventListener('htmx:load', (event) => {
  initComponentNameGuidance(event.target);
});
