import { subscribeToKeyDownEvent, subscribeToPointerEvents } from "./input";

export function init() {
    subscribeToKeyDownEvent();
    subscribeToPointerEvents();
}
