import { useSyncExternalStore } from "react";

/** 1024px 未満を狭い画面とみなす。 */
const query = "(max-width: 1023.98px)";

function subscribe(onChange: () => void) {
  const mql = window.matchMedia?.(query);
  mql?.addEventListener("change", onChange);
  return () => mql?.removeEventListener("change", onChange);
}

/** 画面幅が 1024px 未満なら true。幅が変わると更新される。 */
export function useIsNarrow(): boolean {
  return useSyncExternalStore(subscribe, () => window.matchMedia?.(query).matches ?? false);
}
