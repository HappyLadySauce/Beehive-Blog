import { BookOpen, FileText, FolderKanban, Globe2, Home, MapPin, NotebookTabs, Rss, ShieldCheck, Tags } from "lucide-react";
import Link from "next/link";
import Image from "next/image";

import type { SiteAuthor, SiteSettingsPublic, SiteStats } from "@/lib/api/types";

type FooterLink = {
  href: string;
  label: string;
};

const footerGroups: Array<{ title: string; links: FooterLink[] }> = [
  {
    title: "服务",
    links: [
      { href: "/sitemap.xml", label: "站点地图" },
      { href: "/posts", label: "公开文章" },
      { href: "/notes", label: "公开笔记" }
    ]
  },
  {
    title: "内容",
    links: [
      { href: "/posts", label: "文章" },
      { href: "/notes", label: "笔记" },
      { href: "/projects", label: "项目" }
    ]
  },
  {
    title: "导航",
    links: [
      { href: "/", label: "首页" },
      { href: "/login", label: "登录" },
      { href: "/studio", label: "Studio" }
    ]
  },
  {
    title: "协议",
    links: [
      { href: "/robots.txt", label: "Robots" },
      { href: "/sitemap.xml", label: "索引" },
      { href: "/", label: "版权协议" }
    ]
  }
];

export function AnzhiyuFooter({
  author,
  site,
  stats
}: {
  author: SiteAuthor;
  site?: SiteSettingsPublic;
  stats: SiteStats;
}) {
  const siteName = site?.name?.trim() || "Beehive";
  const authorName = author.name?.trim() || siteName;
  const authorInitial = authorName.slice(0, 1).toUpperCase();
  const authorAvatarURL = author.avatar_url?.trim();
  const siteURL = site?.url?.trim();
  const authorWebsite = author.website?.trim();
  const footerText = site?.footer_text?.trim();
  const icpBeian = site?.icp_beian?.trim();
  const policeBeian = site?.police_beian?.trim();

  return (
    <footer className="anz-footer">
      <div className="anz-footer__inner">
        <div className="anz-footer__socials" aria-label="站点快捷入口">
          <Link href="/" aria-label="首页">
            <Home aria-hidden size={18} />
          </Link>
          <Link href="/posts" aria-label="文章">
            <FileText aria-hidden size={18} />
          </Link>
          <Link href="/notes" aria-label="笔记">
            <NotebookTabs aria-hidden size={18} />
          </Link>
          <Link href="/projects" aria-label="项目">
            <FolderKanban aria-hidden size={18} />
          </Link>
          <Link href="/sitemap.xml" aria-label="站点地图">
            <Rss aria-hidden size={18} />
          </Link>
          {authorWebsite ? (
            <a href={authorWebsite} target="_blank" rel="noreferrer" aria-label="个人网站">
              <Globe2 aria-hidden size={18} />
            </a>
          ) : null}
        </div>

        <div className="anz-footer__profile">
          <div className="anz-footer__avatar">
            {authorAvatarURL ? <Image alt={authorName} height={58} src={authorAvatarURL} unoptimized width={58} /> : authorInitial}
          </div>
          <div>
            <strong>{authorName}</strong>
            <span>{author.description || site?.subtitle || "个人博客与知识中台"}</span>
          </div>
        </div>

        <div className="anz-footer__columns">
          <section className="anz-footer__brand">
            <h2>{siteName}</h2>
            {site?.description ? <p>{site.description}</p> : null}
            {author.bio ? <p>{author.bio}</p> : null}
            <div className="anz-footer__meta">
              {author.location ? (
                <span>
                  <MapPin aria-hidden size={14} />
                  {author.location}
                </span>
              ) : null}
              {siteURL ? (
                <a href={siteURL} target="_blank" rel="noreferrer">
                  {siteURL}
                </a>
              ) : null}
              {authorWebsite ? (
                <a href={authorWebsite} target="_blank" rel="noreferrer">
                  个人网站
                </a>
              ) : null}
            </div>
          </section>

          {footerGroups.map((group) => (
            <nav className="anz-footer__group" key={group.title} aria-label={group.title}>
              <h3>{group.title}</h3>
              {group.links.map((link) => (
                <Link href={link.href} key={`${group.title}-${link.href}-${link.label}`}>
                  {link.label}
                </Link>
              ))}
            </nav>
          ))}
        </div>

        <div className="anz-footer__stats" aria-label="站点统计">
          <span>
            <BookOpen aria-hidden size={15} />
            文章 {stats.articles}
          </span>
          <span>
            <NotebookTabs aria-hidden size={15} />
            笔记 {stats.notes}
          </span>
          <span>
            <FolderKanban aria-hidden size={15} />
            项目 {stats.projects}
          </span>
          <span>
            <Tags aria-hidden size={15} />
            标签 {stats.tags}
          </span>
        </div>

        <div className="anz-footer__bottom">
          <span>© 2020 - {new Date().getFullYear()} By {authorName}</span>
          {icpBeian ? <span>{icpBeian}</span> : null}
          {policeBeian ? <span>{policeBeian}</span> : null}
          {footerText ? <span>{footerText}</span> : null}
          <span>
            <ShieldCheck aria-hidden size={14} />
            所有业务正常
          </span>
        </div>
      </div>
    </footer>
  );
}
