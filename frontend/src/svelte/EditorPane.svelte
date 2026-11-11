<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    label = 'Markdown 編輯器',
    heading = 'MARKDOWN',
    fontSize = 14,
    minFontSize = 10,
    maxFontSize = 32,
    onzoomout,
    onzoomin,
    onzoomreset,
    children,
  }: {
    label?: string;
    heading?: string;
    fontSize?: number;
    minFontSize?: number;
    maxFontSize?: number;
    onzoomout: () => void;
    onzoomin: () => void;
    onzoomreset: () => void;
    children: Snippet;
  } = $props();
</script>

<section class="pane editor-pane" aria-label={label}>
  <div class="pane-heading">
    <span>{heading}</span>

    <div class="font-controls" role="group" aria-label="編輯器文字大小">
      <button type="button" title="縮小文字" aria-label="縮小文字" disabled={fontSize <= minFontSize} onclick={onzoomout}>−</button>
      <span class="font-size">{fontSize}px</span>
      <button type="button" title="放大文字" aria-label="放大文字" disabled={fontSize >= maxFontSize} onclick={onzoomin}>+</button>
      <button type="button" title="重設文字大小" onclick={onzoomreset}>重設</button>
    </div>
  </div>

  <div class="pane-body">
    {@render children()}
  </div>
</section>

<style>
  .editor-pane {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    background: #102a35;
  }

  .pane-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex: 0 0 auto;
    gap: 12px;
    min-height: 32px;
    padding: 4px 14px;
    border-bottom: 1px solid #dce1e7;
    background: #f1f5f9;
    color: #64748b;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 1px;
  }

  .font-controls {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: 6px;
    letter-spacing: normal;
  }

  .font-size {
    min-width: 38px;
    text-align: center;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }

  .font-controls button {
    appearance: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 26px;
    height: 24px;
    padding: 0 6px;
    border: 1px solid #cbd5e1;
    border-radius: 4px;
    background: #ffffff;
    color: #334155;
    font-family: inherit;
    font-size: 12px;
    cursor: pointer;
  }

  .font-controls button:hover:not(:disabled) {
    background: #e2e8f0;
    color: #0f172a;
  }

  .font-controls button:disabled {
    border-color: #e2e8f0;
    background: #f1f5f9;
    color: #94a3b8;
    cursor: not-allowed;
  }

  .font-controls button:focus-visible {
    outline: 2px solid #2563eb;
    outline-offset: 2px;
  }

  .pane-body {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }
</style>
