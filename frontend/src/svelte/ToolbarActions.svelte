<script lang="ts">
  let {
    disabled = false,
    onopen,
    onsave,
    onsaveas,
    onexport,
  }: {
    disabled?: boolean;
    onopen?: () => void;
    onsave?: () => void;
    onsaveas?: () => void;
    onexport?: () => void;
  } = $props();
</script>

<div class="toolbar-actions">
  <button type="button" disabled={disabled} onclick={() => onopen?.()}>開啟</button>
  <button type="button" disabled={disabled} onclick={() => onsave?.()}>儲存</button>
  <button type="button" disabled={disabled} onclick={() => onsaveas?.()}>另存為</button>
  <button type="button" disabled={disabled} onclick={() => onexport?.()}>匯出 HTML</button>
</div>

<style>
  .toolbar-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .toolbar-actions button {
    appearance: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 32px;
    padding: 6px 12px;
    border: 1px solid #cbd5e1;
    border-radius: 6px;

    background: #ffffff;
    color: #334155;

    font-family: inherit;
    font-size: 13px;
    font-weight: 500;
    line-height: 1.4;
    white-space: nowrap;

    opacity: 1;
    cursor: pointer;

    transition:
      background-color 0.15s ease,
      border-color 0.15s ease,
      color 0.15s ease;
  }

  .toolbar-actions button:hover:not(:disabled) {
    border-color: #94a3b8;
    background: #f1f5f9;
    color: #0f172a;
  }

  .toolbar-actions button:active:not(:disabled) {
    border-color: #64748b;
    background: #e2e8f0;
    color: #0f172a;
  }

  .toolbar-actions button:focus-visible {
    outline: 2px solid #2563eb;
    outline-offset: 2px;
  }

  .toolbar-actions button:disabled {
    border-color: #e2e8f0;
    background: #f1f5f9;
    color: #64748b;
    opacity: 1;
    cursor: not-allowed;
  }
</style>
