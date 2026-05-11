# Studio 布局约定

> 文件名沿用 `studio-layout-implementation-plan.md`，内容为**持续有效的约定与修改指南**，与实施阶段无关。

| 属性 | 说明 |
|------|------|
| **范围** | Beehive-Blog **Studio（`/studio/*`）** 主壳：流式占宽、主区内响应式栅格、与全站顶栏视觉协调；**不改变**「Client-heavy、弱 SEO」的双轨定位。 |
| **依赖** | 须与 [React 前端技术架构：强 SSR / SEO](./react-ssr-seo-architecture.md) **§3.3（样式与可访问性）**、**§4.2（双轨渲染）** 一致。 |
| **实现锚点** | [`StudioLayout.tsx`](../../ui/components/studio/StudioLayout.tsx)、[`Studio.module.css`](../../ui/components/studio/Studio.module.css)；全站顶栏与公开页留白见 [`globals.css`](../../ui/app/globals.css)。 |

---

## 1. 与架构文档对齐的硬性约定

### 1.1 样式体系

- 仓库级为 **CSS Modules** 或 **Tailwind** **二选一**，避免三种以上样式体系并行。
- **Studio 仅允许在 CSS Modules 内**扩展布局与主题相关样式；使用 **CSS 变量、Grid、`@media`、`@container`**，不为此单独引入 MUI / Chakra / Mantine 等整套 UI 库作为栅格主力。
- **不为 Studio 单独叠加 Tailwind**；若整站迁移到 Tailwind，再统一替换，而非局部混用。

### 1.2 Studio 与 SSR / SEO

- Studio 为 **Client-heavy**，**不追求**可索引首屏；路由级 `noindex` 等策略见架构文档与 `robots` 配置。
- **布局与换列不得依赖 RSC**；与公开页的 SSR / SEO 数据路径保持解耦。

### 1.3 结构性分支与 JS

- 若需「窄屏抽屉侧栏 / 宽屏固定侧栏」等**无法用 CSS 单独表达**的分支，可用 **`useMediaQuery` 类 Hook** 做小范围判断。
- **默认定价仍以 CSS 为主**（容器查询 + 媒体查询），避免多余 hydration 与首屏闪烁。

---

## 2. 布局与响应式约定（维护时须遵守）

1. **主壳**：主内容区应 **占满可用宽度**（例如 `.layout` 使用 `width: 100%`、`max-width: none`），与 `.site-header` 的水平 padding 策略协调，避免「中间一条孤岛」。
2. **主区内断点**：在承载卡片栅格的区域使用 **`container-type: inline-size`**（及稳定的 `container-name`），用 **`@container`** 驱动 `.metricGrid`、`.workGrid` 等列数变化。
3. **整页断点**：侧栏与主区 **叠成单栏** 等整页结构变化保留 **`@media`**；与 `@container` 分工明确——**整页** vs **主内容区宽度**，勿混用同一套断点语义。
4. **间距与尺寸**：使用 **Studio 专用 CSS 变量**（如 `--studio-pad-x`、`--studio-gap`、`--studio-sidebar-w`）表达 padding / gap / 侧栏宽，**避免魔法数字散落**。
5. **紧凑度**：调整 `min-height`、内边距时须兼顾 **可读性与触控目标**，改动后做关键视口目测。

---

## 3. 修改 Studio 壳层时的检查项

| 检查项 | 说明 |
|--------|------|
| **视口** | 在约 **1280 / 1024 / 900 / 600** px 下确认指标卡列数、工作区列数、侧栏折叠与现网一致或符合新设计。 |
| **测试** | 运行 `StudioOverview`、`StudioSidebar` 等与布局相关的测试；**仅改 CSS** 时通常无需改测，**改 DOM 结构**时须评估快照与 a11y。 |

---

## 4. 与公开页可选对齐

若希望 Public 与 Studio **水平留白 / max-width** 观感一致，可在 `globals.css` 的 `.page` 等与 Studio 变量策略对齐；属产品决策，**非架构强制**。

---

## 5. 兼容性与风险

| 主题 | 约定 |
|------|------|
| **容器查询** | 以现代浏览器为准；若必须支持极旧环境，可退化为更细的 `@media`（会牺牲「随主区宽度」语义）。 |
| **触控与可读性** | 压缩纵向空间时保留合理 `min-height` 与字号，避免为紧凑牺牲可用性。 |

---

*架构或框架升级后，请复核与 [react-ssr-seo-architecture.md](./react-ssr-seo-architecture.md) §3.3、§4.2 是否仍一致。*
