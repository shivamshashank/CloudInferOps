export default function Empty({
  title,
  copy,
}: {
  title: string;
  copy: string;
}) {
  return (
    <div className="empty">
      <div>◇</div>
      <strong>{title}</strong>
      <p>{copy}</p>
    </div>
  );
}
