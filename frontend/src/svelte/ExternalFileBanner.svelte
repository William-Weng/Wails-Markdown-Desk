<script lang="ts">
  type ExternalFileBannerVariant = 'error' | 'conflict';

  let {
    variant = 'conflict',
    message = '',
    disabled = false,
    confirmLabel = '重新載入',
    cancelLabel = '取消',
    onconfirm,
    oncancel,
  }: {
    variant?: ExternalFileBannerVariant;
    message?: string;
    disabled?: boolean;
    confirmLabel?: string;
    cancelLabel?: string;
    onconfirm?: () => void;
    oncancel?: () => void;
  } = $props();
</script>

{#if message}
  <div class="external-banner" class:external-error={variant === 'error'} role={variant === 'error' ? 'alert' : 'status'}>
    <span class="external-message">{message}</span>
    {#if onconfirm || oncancel}
      <div class="external-actions">
        {#if onconfirm}
          <button type="button" disabled={disabled} onclick={onconfirm}>{confirmLabel}</button>
        {/if}
        {#if oncancel}
          <button type="button" disabled={disabled} onclick={oncancel}>{cancelLabel}</button>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .external-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 14px;
    border: 1px solid rgba(255, 193, 7, 0.35);
    border-radius: 8px;
    background: rgba(255, 193, 7, 0.12);
    color: #ffe08a;
  }

  .external-banner.external-error {
    border-color: rgba(255, 107, 107, 0.35);
    background: rgba(255, 107, 107, 0.12);
    color: #ffb4b4;
  }

  .external-message {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 13px;
  }

  .external-actions {
    display: flex;
    flex-shrink: 0;
    gap: 8px;
  }

  .external-actions button {
    padding: 4px 10px;
    border: 1px solid currentColor;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font-size: 12px;
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }

  .external-actions button:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
  }

  .external-actions button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
