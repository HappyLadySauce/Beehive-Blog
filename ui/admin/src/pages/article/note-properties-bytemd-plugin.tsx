import type { BytemdPlugin } from 'bytemd';
import { createRoot } from 'react-dom/client';
import ObsidianNoteProperties from './ObsidianNoteProperties';
import { useArticleNotePropsStore } from './articleNotePropsStore';

function NotePropertiesHost() {
  const show = useArticleNotePropsStore((s) => s.showNoteProperties);
  if (!show) {
    return null;
  }
  return <ObsidianNoteProperties />;
}

/**
 * 在 ByteMD 左侧编辑区、右侧预览区顶部各注入一块「笔记属性」UI（与 Obsidian 分栏一致）。
 */
export function createNotePropertiesBytemdPlugin(): BytemdPlugin {
  return {
    editorEffect(ctx) {
      const editor = ctx.root.querySelector('.bytemd-editor');
      if (!editor) {
        return undefined;
      }
      const el = document.createElement('div');
      el.className = 'beehive-note-properties-root';
      editor.insertBefore(el, editor.firstChild);
      const root = createRoot(el);
      root.render(<NotePropertiesHost />);
      return () => {
        root.unmount();
        el.remove();
      };
    },
    viewerEffect(ctx) {
      const parent = ctx.markdownBody.parentElement;
      if (!parent) {
        return undefined;
      }
      const el = document.createElement('div');
      el.className = 'beehive-note-properties-root';
      parent.insertBefore(el, ctx.markdownBody);
      const root = createRoot(el);
      root.render(<NotePropertiesHost />);
      return () => {
        root.unmount();
        el.remove();
      };
    },
  };
}
