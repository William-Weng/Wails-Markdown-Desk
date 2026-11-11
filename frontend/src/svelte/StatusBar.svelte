<script lang="ts">
  type StatusBarState = 'busy' | 'dirty' | 'saved' | 'empty';

  let {
    lineCount = 0,
    characterCount = 0,
    cursorLine = 1,
    cursorColumn = 1,
    filePath = '',
    state = 'empty',
  }: {
    lineCount?: number;
    characterCount?: number;
    cursorLine?: number;
    cursorColumn?: number;
    filePath?: string;
    state?: StatusBarState;
  } = $props();

  const statusText = $derived.by(() => {
    switch (state) {
      case 'busy': return '處理中…';
      case 'dirty': return '未儲存變更';
      case 'saved': return filePath;
      default: return '未命名文件';
    }
  });
</script>

<footer class="statusbar">
  <span>{lineCount} 行</span>
  <span>{characterCount} 字元</span>
  <span>Ln {cursorLine}, Col {cursorColumn}</span>
  <span class="status-right" title={filePath}>{statusText}</span>
</footer>

<style>
  .statusbar {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 6px 14px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    background: #0c222c;
    color: #8fa3ad;
    font-size: 12px;
  }

  .status-right {
    min-width: 0;
    margin-left: auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
