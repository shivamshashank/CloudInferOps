import { tone } from "../utils";

export default function Status({ value }: { value: string }) {
  return (
    <span className={`status ${tone(value)}`}>
      <i />
      {value}
    </span>
  );
}
