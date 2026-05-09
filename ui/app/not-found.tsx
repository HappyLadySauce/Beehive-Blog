import Link from "next/link";

export default function NotFound() {
  return (
    <main className="page">
      <section className="page-title">
        <p className="eyebrow">404</p>
        <h1>页面不存在</h1>
        <p>这个地址没有对应的公开内容。</p>
        <Link className="primary-button" href="/posts">返回文章列表</Link>
      </section>
    </main>
  );
}
