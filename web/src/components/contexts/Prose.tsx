/** Renders whyspec section text: bullet lines become a list, others paragraphs. */
export function Prose({ text }: { text: string }) {
  if (!text || !text.trim()) {
    return <p className="text-[13px] italic text-ink-muted/60">Chưa ghi nhận.</p>;
  }
  const lines = text.split("\n").map((l) => l.trimEnd());
  const blocks: { type: "p" | "li"; text: string }[] = [];
  for (const line of lines) {
    const t = line.trim();
    if (!t) continue;
    if (t.startsWith("- ") || t.startsWith("* ")) {
      blocks.push({ type: "li", text: t.slice(2) });
    } else {
      blocks.push({ type: "p", text: t });
    }
  }

  const out: React.ReactNode[] = [];
  let bucket: string[] = [];
  const flush = (key: number) => {
    if (bucket.length) {
      out.push(
        <ul key={`ul-${key}`} className="my-1 flex flex-col gap-1.5">
          {bucket.map((b, i) => (
            <li key={i} className="flex gap-2.5 text-[13.5px] leading-relaxed text-ink/85">
              <span className="mt-2 h-1 w-1 shrink-0 rounded-full bg-accent" />
              <span>{b}</span>
            </li>
          ))}
        </ul>
      );
      bucket = [];
    }
  };
  blocks.forEach((b, i) => {
    if (b.type === "li") {
      bucket.push(b.text);
    } else {
      flush(i);
      out.push(
        <p key={`p-${i}`} className="text-[13.5px] leading-relaxed text-ink/85">
          {b.text}
        </p>
      );
    }
  });
  flush(blocks.length);

  return <div className="flex flex-col gap-2.5 max-w-[68ch]">{out}</div>;
}
