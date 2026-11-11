<script lang="ts">
  type ViewMode = 'code' | 'html' | 'split';

  let {
    value = $bindable(),
    disabled = false,
    onchange,
  }: {
    value?: ViewMode;
    disabled?: boolean;
    onchange?: (mode: ViewMode) => void;
  } = $props();

  const options: Array<{
    value: ViewMode;
    label: string;
  }> = [
    { value: 'code', label: 'Code' },
    { value: 'html', label: 'HTML' },
    { value: 'split', label: 'Code / HTML' },
  ];

  function select(mode: ViewMode) {
    if (disabled || value === mode) return;

    value = mode;
    onchange?.(mode);
  }
</script>

<div class="view-switch" role="group" aria-label="檢視模式">
  {#each options as option (option.value)}
    <button type="button" class:active={value === option.value} aria-pressed={value === option.value} disabled={disabled} onclick={() => select(option.value)}>{option.label}</button>
  {/each}
</div>

<style>
  .view-switch {
    display: inline-flex;
    gap: 4px;
    padding: 3px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
  }

  .view-switch button {
    min-width: 72px;
    padding: 5px 10px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: #b8c7d1;
    font-size: 12px;
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }

  .view-switch button:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    color: #e6f1f6;
  }

  .view-switch button.active {
    background: #35596d;
    color: #ffffff;
  }

  .view-switch button:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
</style>
