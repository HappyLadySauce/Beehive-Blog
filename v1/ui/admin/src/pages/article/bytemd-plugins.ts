import type { BytemdPlugin } from 'bytemd';
import gfm from '@bytemd/plugin-gfm';
import highlight from '@bytemd/plugin-highlight';
import rehypePreLanguage from 'rehype-pre-language';

/**
 * ByteMD 插件链：GFM、highlight.js，最后将 fenced 语言的 class 同步到 pre[data-language]（供角标 CSS 使用）。
 */
const rehypePreLanguagePlugin: BytemdPlugin = {
  rehype: (processor) => processor.use(rehypePreLanguage, 'data-language'),
};

export const articleBytemdPlugins: BytemdPlugin[] = [gfm(), highlight(), rehypePreLanguagePlugin];
