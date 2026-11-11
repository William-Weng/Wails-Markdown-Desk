<script lang="ts">
  let {
    title = 'MarkdownDesk',
    fileName = 'Untitled.md',
    filePath = '',
    isDirty = false,
  }: {
    title?: string;
    fileName?: string;
    filePath?: string;
    isDirty?: boolean;
  } = $props();
</script>

<div class="toolbar-document" title={filePath}>
  <span class="app-title">{title}</span>
  <span class="document-name">
    {fileName}{isDirty ? ' •' : ''}
  </span>
</div>

<style>
  .toolbar-document {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .app-title {
    color: #9fb6c2;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    white-space: nowrap;
  }

  .document-name {
    min-width: 0;
    overflow: hidden;
    color: #e6f1f6;
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
