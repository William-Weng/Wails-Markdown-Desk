<script lang="ts">
  let {
    label = 'Markdown 預覽',
    heading = 'PREVIEW',
    html = '',
  }: {
    label?: string;
    heading?: string;
    html?: string;
  } = $props();
</script>

<section class="pane preview-pane" aria-label={label}>
  <div class="pane-heading">{heading}</div>
  <div class="preview-scroll">
    <article class="markdown-body">{@html html}</article>
  </div>
</section>

<style>
  .preview-pane {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    background: #ffffff;
    color: #20252b;
  }

  .pane-heading {
    display: flex;
    align-items: center;
    flex: 0 0 32px;
    padding: 0 14px;
    border-bottom: 1px solid #dce1e7;
    background: #f1f5f9;
    color: #64748b;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 1px;
  }

  .preview-scroll {
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: auto;
  }
</style>
