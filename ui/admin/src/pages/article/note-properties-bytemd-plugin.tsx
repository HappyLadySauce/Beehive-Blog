import type { BytemdPlugin } from 'bytemd';
import { createRoot } from 'react-dom/client';
import ObsidianNoteProperties from './ObsidianNoteProperties';
import { useArticleNotePropsStore } from './articleNotePropsStore';

function NotePropertiesHost({ readOnly }: { readOnly?: boolean }) {
  const show = useArticleNotePropsStore((s) => s.showNoteProperties);
  if (!show) {
    return null;
  }
  return <ObsidianNoteProperties readOnly={readOnly} />;
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
      root.render(<NotePropertiesHost readOnly={false} />);
      return () => {
        // ByteMD 可能在父组件仍处渲染阶段时调用 cleanup；同步 root.unmount() 会触发 React 警告，故推迟到下一 macrotask。
        const r = root;
        const node = el;
        setTimeout(() => {
          try {
            r.unmount();
          } catch {
            /* ignore */
          }
          node.remove();
        }, 0);
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
      root.render(<NotePropertiesHost readOnly />);
      return () => {
        const r = root;
        const node = el;
        setTimeout(() => {
          try {
            r.unmount();
          } catch {
            /* ignore */
          }
          node.remove();
        }, 0);
      };
    },
  };
}
