import { useEffect, useSyncExternalStore } from "react";

// 画面は少ないので、ルーティングのライブラリは入れずに History API で URL を切り替える。

function subscribe(onChange: () => void) {
  window.addEventListener("popstate", onChange);
  return () => window.removeEventListener("popstate", onChange);
}

/** 今の URL のパス。ブラウザの戻る・進むでも更新される。 */
export function usePath(): string {
  return useSyncExternalStore(subscribe, () => window.location.pathname);
}

/** path へ移る（履歴に積む）。 */
export function navigate(path: string) {
  if (path === window.location.pathname) return;
  window.history.pushState(null, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

/** `/questions/{id}` なら id を返す。 */
export function questionIdFromPath(path: string): number | undefined {
  const m = /^\/questions\/(\d+)\/?$/.exec(path);
  return m ? Number(m[1]) : undefined;
}

// 画面を離れると消える入力があるか。ヘッダーなど、その画面の外から移るときにも確認できるようにする。
let unsaved = false;

/** active の間、confirmLeave で確認を出すようにする（解答中に空欄を埋めているときなど）。 */
export function useLeaveGuard(active: boolean) {
  useEffect(() => {
    unsaved = active;
    return () => {
      unsaved = false;
    };
  }, [active]);
}

/** 消える入力があれば確認を出し、移ってよいかを返す。 */
export function confirmLeave(): boolean {
  return !unsaved || window.confirm("解答が消えますがよいですか？");
}
