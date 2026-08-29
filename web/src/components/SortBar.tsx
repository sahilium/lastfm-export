import type { SortOption } from '../lib/api';

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'name_asc', label: 'A–Z' },
  { value: 'name_desc', label: 'Z–A' },
  { value: 'scrobbles_desc', label: 'Most played' },
  { value: 'scrobbles_asc', label: 'Least played' },
  { value: 'recent_desc', label: 'Recently played' },
  { value: 'recent_asc', label: 'Oldest played' },
];

interface SortBarProps {
  value: SortOption;
  onChange: (sort: SortOption) => void;
}

export default function SortBar({ value, onChange }: SortBarProps) {
  return (
    <div className="sort-bar">
      <label className="sort-label">Sort by</label>
      <select
        className="sort-select"
        value={value}
        onChange={e => onChange(e.target.value as SortOption)}
      >
        {SORT_OPTIONS.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  );
}
