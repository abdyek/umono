class ComponoErrorBase extends HTMLElement {
  constructor(isBlock) {
    super()
    const shadow = this.attachShadow({ mode: "closed" })

    shadow.innerHTML = `
<style>
  :host {
    font-family: monospace;
    color: #000;
    font-size: 1rem;
    font-weight: normal;
  }

  .wrapper {
    border: 2px solid #ff0000;
    background: #fff0f0;
    padding: 6px 8px;
  }

  .inline {
    display: inline-block;
  }

  .block {
    display: block;
    margin: 1em 0;
  }

  .title {
    font-weight: bold;
    background: #ff0000;
    color: #fff;
    padding: 2px 4px;
    margin-bottom: 4px;
    display: inline-block;
  }

  .description {
    font-size: 0.9em;
    line-height: 1.3;
  }
</style>

<div class="wrapper ${isBlock ? "block" : "inline"}">
  <div class="title">
    <slot name="title"></slot>
  </div>
  <div class="description">
    <slot name="description"></slot>
  </div>
</div>
`
  }
}

class ComponoErrorInline extends ComponoErrorBase {
  constructor() {
    super(false)
  }
}

class ComponoErrorBlock extends ComponoErrorBase {
  constructor() {
    super(true)
  }
}

customElements.define("compono-error-inline", ComponoErrorInline)
customElements.define("compono-error-block", ComponoErrorBlock)
