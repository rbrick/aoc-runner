type Props = {
  earned: number;
  total?: number;
};

export function Stars({ earned, total = 2 }: Props) {
  return (
    <span className="stars" aria-label={`${earned} of ${total} stars`} title={`${earned} of ${total} stars`}>
      {Array.from({ length: total }, (_, index) => (
        <span key={index} className={index < earned ? "star star-earned" : "star"} aria-hidden="true">
          ★
        </span>
      ))}
    </span>
  );
}
