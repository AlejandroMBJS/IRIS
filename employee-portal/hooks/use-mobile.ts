/**
 * @file hooks/use-mobile.ts
 * @description Responsive design hook that detects mobile viewport.
 *   Uses matchMedia API for efficient breakpoint detection.
 *
 * USER PERSPECTIVE:
 *   - Enables mobile-optimized layouts automatically
 *   - Responsive UI adapts to screen size changes
 *
 * DEVELOPER GUIDELINES:
 *   OK to modify: MOBILE_BREAKPOINT value (currently 768px)
 *   CAUTION: SSR compatibility (returns undefined initially)
 *   DO NOT modify: matchMedia event listener pattern
 *
 * KEY EXPORTS:
 *   - useIsMobile: Returns true if viewport < 768px
 */
import * as React from "react"

const MOBILE_BREAKPOINT = 768

export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState<boolean | undefined>(undefined)

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`)
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    }
    mql.addEventListener("change", onChange)
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT)
    return () => mql.removeEventListener("change", onChange)
  }, [])

  return !!isMobile
}
