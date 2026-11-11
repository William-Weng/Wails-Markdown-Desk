<script lang="ts">
  let {
    message = '',
    dismissLabel = '關閉',
    ondismiss,
  }: {
    message?: string;
    dismissLabel?: string;
    ondismiss: () => void;
  } = $props();
</script>

{#if message}
  <div class="export-banner" role="status">
    <span class="export-message">{message}</span>
    <button type="button" onclick={ondismiss}>{dismissLabel}</button>
  </div>
{/if}

<style>
  .export-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex: 0 0 auto;
    gap: 12px;
    padding: 10px 14px;
    border-top: 1px solid #bbf7d0;
    background: #f0fdf4;
    color: #166534;
    font-size: 13px;
    line-height: 1.5;
  }

  .export-message {
    min-width: 0;
    overflow-wrap: anywhere;
    white-space: pre-wrap;
  }

  .export-banner button {
    appearance: none;
    flex-shrink: 0;
    padding: 4px 10px;
    border: 1px solid #86efac;
    border-radius: 6px;
    background: #ffffff;
    color: #166534;
    font-family: inherit;
    font-size: 12px;
    line-height: 1.5;
    cursor: pointer;

    transition:
      background-color 0.15s ease,
      border-color 0.15s ease;
  }

  .export-banner button:hover {
    border-color: #4ade80;
    background: #dcfce7;
  }

  .export-banner button:focus-visible {
    outline: 2px solid #15803d;
    outline-offset: 2px;
  }
</style>
