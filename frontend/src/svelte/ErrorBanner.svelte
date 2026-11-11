<script lang="ts">
  let {
    message = '',
    dismissLabel = '關閉',
    ondismiss,
  }: {
    message?: string;
    dismissLabel?: string;
    ondismiss?: () => void;
  } = $props();
</script>

{#if message}
  <div class="error-banner" role="alert">
    <span class="error-message">{message}</span>
    {#if ondismiss}
      <button type="button" onclick={ondismiss}>{dismissLabel}</button>
    {/if}
  </div>
{/if}

<style>
  .error-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 14px;
    border: 1px solid rgba(255, 107, 107, 0.35);
    border-radius: 8px;
    background: rgba(255, 107, 107, 0.12);
    color: #ffb4b4;
  }

  .error-message {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 13px;
  }

  .error-banner button {
    flex-shrink: 0;
    padding: 4px 10px;
    border: 1px solid rgba(255, 107, 107, 0.35);
    border-radius: 6px;
    background: transparent;
    color: #ffb4b4;
    font-size: 12px;
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }

  .error-banner button:hover {
    background: rgba(255, 107, 107, 0.16);
    color: #ffffff;
  }
</style>
