"use client";

import { MessageCircle, Send } from "lucide-react";
import { FormEvent, useState } from "react";

import { createPublicComment } from "@/lib/api/content";
import type { CommentItem } from "@/lib/api/types";

export function AnzhiyuCommentBox({ comments, contentID }: { comments: CommentItem[]; contentID: number }) {
  const [message, setMessage] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const target = event.currentTarget;
    const form = new FormData(target);
    setSubmitting(true);
    setMessage(null);
    try {
      await createPublicComment(contentID, {
        nickname: String(form.get("nickname") ?? ""),
        email: String(form.get("email") ?? ""),
        website: String(form.get("website") ?? ""),
        body: String(form.get("body") ?? "")
      });
      target.reset();
      setMessage("评论已提交，审核通过后会显示。");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "评论提交失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="anz-comments" id="comments">
      <h2>
        <MessageCircle aria-hidden size={22} />
        评论
        <span>{comments.length}</span>
      </h2>
      <form className="anz-comment-form" onSubmit={onSubmit}>
        <textarea name="body" required maxLength={2000} placeholder="欢迎留下宝贵的建议哟 ~" />
        <div>
          <input name="nickname" required maxLength={80} placeholder="昵称 必填" />
          <input name="email" required maxLength={255} type="email" placeholder="邮箱 必填" />
          <input name="website" maxLength={512} type="url" placeholder="网站 选填" />
        </div>
        <button type="submit" disabled={submitting}>
          <Send aria-hidden size={16} />
          发送
        </button>
        {message ? <p className="anz-comment-form__message">{message}</p> : null}
      </form>
      <div className="anz-comment-list">
        {comments.map((comment) => (
          <article key={comment.id}>
            <strong>{comment.nickname}</strong>
            <time dateTime={comment.created_at}>{new Intl.DateTimeFormat("zh-CN").format(new Date(comment.created_at))}</time>
            <p>{comment.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
