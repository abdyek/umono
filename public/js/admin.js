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

function initMediaAliasGuidance(root = document) {
  const input = findSelfOrDescendant(root, '[data-media-alias-input]');
  const helper = findSelfOrDescendant(root, '[data-media-alias-helper]');
  const helperDot = findSelfOrDescendant(root, '[data-media-alias-helper-dot]');
  const error = findSelfOrDescendant(root, '[data-media-alias-error]');

  if (!input || !helper || !helperDot || !error || input.dataset.mediaAliasBound === 'true') {
    return;
  }

  const mediaAliasPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
  const mode = input.dataset.mediaAliasMode;
  const hasServerError = input.dataset.mediaAliasHasServerError === 'true';
  let hasInteracted = false;

  const updateState = () => {
    const value = input.value.trim();
    const isNeutral = value === '';
    const isValid = mediaAliasPattern.test(value);
    const shouldShow = mode === 'create' || (hasInteracted && !isNeutral);

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

  input.dataset.mediaAliasBound = 'true';
  input.addEventListener('input', () => {
    hasInteracted = true;
    updateState();
  });
  updateState();
}

function findMediaUploadElements(root) {
  const form = findSelfOrDescendant(root, '[data-media-upload-form]');
  if (!form) {
    return null;
  }

  return {
    form,
    fileInput: form.querySelector('#media-file'),
    aliasInput: form.querySelector('#media-alias'),
    storageSelect: form.querySelector('[data-media-storage-select]'),
    progress: form.querySelector('[data-media-upload-progress]'),
    progressLabel: form.querySelector('[data-media-upload-progress-label]'),
    progressValue: form.querySelector('[data-media-upload-progress-value]'),
    progressBar: form.querySelector('[data-media-upload-progress-bar]'),
    submitButton: form.querySelector('button[type="submit"]'),
    serverError: form.querySelector('[data-media-upload-server-error]'),
    aliasError: form.querySelector('[data-media-alias-error]'),
    aliasErrorText: form.querySelector('[data-media-alias-error-text]'),
    aliasHelper: form.querySelector('[data-media-alias-helper]'),
  };
}

function setMediaUploadProgress(elements, value, label) {
  if (!elements.progress || !elements.progressValue || !elements.progressBar || !elements.progressLabel) {
    return;
  }

  const percent = Math.max(0, Math.min(100, Math.round(value)));
  elements.progress.classList.remove('hidden');
  elements.progressValue.textContent = `${percent}%`;
  elements.progressBar.style.width = `${percent}%`;
  if (label) {
    elements.progressLabel.textContent = label;
  }
}

function hideMediaUploadProgress(elements) {
  if (!elements.progress) {
    return;
  }

  elements.progress.classList.add('hidden');
}

function setMediaUploadBusy(elements, busy) {
  elements.form.classList.toggle('pointer-events-none', busy);
  elements.form.classList.toggle('opacity-90', busy);
  if (elements.submitButton) {
    elements.submitButton.disabled = busy;
  }
}

function clearMediaUploadError(elements) {
  if (elements.serverError) {
    elements.serverError.classList.add('hidden');
    elements.serverError.textContent = '';
  }
  if (elements.aliasError && elements.aliasErrorText) {
    elements.aliasError.classList.add('hidden');
    elements.aliasError.classList.remove('flex');
    elements.aliasErrorText.textContent = '';
  }
  if (elements.aliasInput) {
    elements.aliasInput.dataset.mediaAliasHasServerError = 'false';
  }
}

function setMediaUploadError(elements, message, aliasError = false) {
  if (aliasError && elements.aliasError && elements.aliasErrorText) {
    elements.aliasError.classList.remove('hidden');
    elements.aliasError.classList.add('flex');
    elements.aliasErrorText.textContent = message;
    if (elements.aliasHelper) {
      elements.aliasHelper.classList.add('hidden');
    }
    if (elements.aliasInput) {
      elements.aliasInput.dataset.mediaAliasHasServerError = 'true';
      elements.aliasInput.classList.remove('border-neutral-700', 'border-emerald-500/40', 'border-amber-500/40');
      elements.aliasInput.classList.add('border-red-500/50');
    }
    return;
  }

  if (elements.serverError) {
    elements.serverError.classList.remove('hidden');
    elements.serverError.textContent = message;
  }
}

async function sha256Hex(file) {
  const buffer = await file.arrayBuffer();
  const digest = await crypto.subtle.digest('SHA-256', buffer);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
}

function uploadFileWithXHR(url, headers, file, onProgress) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('PUT', url, true);

    Object.entries(headers || {}).forEach(([key, value]) => {
      xhr.setRequestHeader(key, value);
    });
    if (!headers || (!('Content-Type' in headers) && !('content-type' in headers))) {
      xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream');
    }

    xhr.upload.addEventListener('progress', (event) => {
      if (!event.lengthComputable) {
        return;
      }
      onProgress((event.loaded / event.total) * 100);
    });

    xhr.onerror = () => reject(new Error('upload_failed'));
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
        return;
      }
      reject(new Error('upload_failed'));
    };

    xhr.send(file);
  });
}

function applyMediaContentResponse(html, pushUrl) {
  const parser = new DOMParser();
  const nextDocument = parser.parseFromString(html, 'text/html');
  const nextContent = nextDocument.querySelector('#media-content');
  const currentContent = document.querySelector('#media-content');

  if (!nextContent || !currentContent) {
    throw new Error('Upload failed.');
  }

  currentContent.outerHTML = nextContent.outerHTML;

  if (pushUrl && pushUrl !== 'false') {
    window.history.pushState({}, '', pushUrl);
  }

  const replacedContent = document.querySelector('#media-content');
  if (replacedContent) {
    htmx.process(replacedContent);
    initComponentNameGuidance(replacedContent);
    initMediaAliasGuidance(replacedContent);
    initMediaUpload(replacedContent);
  }
}

async function submitLocalMediaUpload(elements) {
  const formData = new FormData(elements.form);
  const response = await fetch('/admin/media', {
    method: 'POST',
    headers: {
      'HX-Request': 'true',
      'HX-Target': 'media-content',
      'X-Umono-Media-Upload': 'true',
    },
    body: formData,
  });

  const contentType = response.headers.get('Content-Type') || '';
  if (!response.ok) {
    if (contentType.includes('application/json')) {
      const uploadError = await response.json();
      throw { message: uploadError.error || 'Upload failed.', aliasError: uploadError.alias_error === true };
    }
    throw new Error('Upload failed.');
  }

  const html = await response.text();
  applyMediaContentResponse(html, response.headers.get('HX-Push-Url'));
}

function initMediaUpload(root = document) {
  const elements = findMediaUploadElements(root);
  if (!elements || elements.form.dataset.mediaUploadBound === 'true') {
    return;
  }

  elements.form.dataset.mediaUploadBound = 'true';

  elements.form.addEventListener('submit', async (event) => {
    const selectedOption = elements.storageSelect && elements.storageSelect.selectedOptions ? elements.storageSelect.selectedOptions[0] : null;
    const storageType = selectedOption ? selectedOption.dataset.storageType : 'local';
    event.preventDefault();
    event.stopPropagation();
    if (typeof event.stopImmediatePropagation === 'function') {
      event.stopImmediatePropagation();
    }

    const file = elements.fileInput && elements.fileInput.files ? elements.fileInput.files[0] : null;
    if (!file) {
      setMediaUploadError(elements, 'Select a PNG, JPEG, or WEBP image to continue.');
      return;
    }

    clearMediaUploadError(elements);
    setMediaUploadBusy(elements, true);

    try {
      if (storageType !== 's3') {
        setMediaUploadProgress(elements, 20, 'Uploading to local storage…');
        await submitLocalMediaUpload(elements);
        return;
      }

      setMediaUploadProgress(elements, 5, 'Preparing direct upload…');
      const hash = await sha256Hex(file);
      const payload = new URLSearchParams({
        storage_id: elements.storageSelect ? elements.storageSelect.value : '',
        original_name: file.name,
        alias: elements.aliasInput ? elements.aliasInput.value.trim() : '',
        mime_type: file.type || 'application/octet-stream',
        size: `${file.size}`,
        hash,
      });

      const presignResponse = await fetch('/admin/media/presign', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
          'HX-Request': 'true',
        },
        body: payload.toString(),
      });

      const presignData = await presignResponse.json();
      if (!presignResponse.ok) {
        throw { message: presignData.error || 'Upload failed.', aliasError: presignData.alias_error === true };
      }

      setMediaUploadProgress(elements, 15, 'Uploading directly to storage…');
      await uploadFileWithXHR(presignData.url, presignData.headers, file, (progress) => {
        setMediaUploadProgress(elements, progress, 'Uploading directly to storage…');
      });

      setMediaUploadProgress(elements, 100, 'Finalizing media record…');
      const completeResponse = await fetch('/admin/media/complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
          'HX-Request': 'true',
          'HX-Target': 'media-content',
        },
        body: new URLSearchParams({ token: presignData.token }).toString(),
      });

      const contentType = completeResponse.headers.get('Content-Type') || '';
      if (!completeResponse.ok) {
        if (contentType.includes('application/json')) {
          const completeError = await completeResponse.json();
          throw { message: completeError.error || 'Upload failed.', aliasError: completeError.alias_error === true };
        }
        throw new Error('Upload failed.');
      }

      const html = await completeResponse.text();
      applyMediaContentResponse(html, completeResponse.headers.get('HX-Push-Url'));
    } catch (error) {
      hideMediaUploadProgress(elements);
      setMediaUploadBusy(elements, false);
      setMediaUploadError(elements, error.message || 'Upload failed.', error.aliasError === true);
    }
  }, true);
}

document.addEventListener('DOMContentLoaded', () => {
  initComponentNameGuidance();
  initMediaAliasGuidance();
  initMediaUpload();
});

document.addEventListener('htmx:load', (event) => {
  initComponentNameGuidance(event.target);
  initMediaAliasGuidance(event.target);
  initMediaUpload(event.target);
});
