'use client'

import { useState, useRef, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react'

interface HoldToConfirmButtonProps {
  onConfirm: () => void | Promise<void>
  variant: 'approve' | 'decline'
  holdDuration?: number // in milliseconds
  disabled?: boolean
  className?: string
  children?: React.ReactNode
}

export function HoldToConfirmButton({
  onConfirm,
  variant,
  holdDuration = 1000,
  disabled = false,
  className,
  children,
}: HoldToConfirmButtonProps) {
  const [isHolding, setIsHolding] = useState(false)
  const [progress, setProgress] = useState(0)
  const [isExecuting, setIsExecuting] = useState(false)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)
  const startTimeRef = useRef<number>(0)

  const startHold = useCallback(() => {
    if (disabled || isExecuting) return

    setIsHolding(true)
    setProgress(0)
    startTimeRef.current = Date.now()

    intervalRef.current = setInterval(() => {
      const elapsed = Date.now() - startTimeRef.current
      const newProgress = Math.min((elapsed / holdDuration) * 100, 100)
      setProgress(newProgress)

      if (newProgress >= 100) {
        // Clear interval and execute
        if (intervalRef.current) {
          clearInterval(intervalRef.current)
          intervalRef.current = null
        }
        setIsHolding(false)
        executeAction()
      }
    }, 16) // ~60fps
  }, [disabled, isExecuting, holdDuration])

  const cancelHold = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    setIsHolding(false)
    setProgress(0)
  }, [])

  const executeAction = async () => {
    setIsExecuting(true)
    try {
      await onConfirm()
    } finally {
      setIsExecuting(false)
      setProgress(0)
    }
  }

  const isApprove = variant === 'approve'
  const baseColor = isApprove ? 'green' : 'red'

  return (
    <button
      type="button"
      disabled={disabled || isExecuting}
      onMouseDown={startHold}
      onMouseUp={cancelHold}
      onMouseLeave={cancelHold}
      onTouchStart={startHold}
      onTouchEnd={cancelHold}
      onTouchCancel={cancelHold}
      className={cn(
        'relative overflow-hidden rounded-md px-3 py-1.5 text-sm font-medium transition-all duration-150',
        'flex items-center gap-1.5 select-none',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        // Base styles
        isApprove
          ? 'bg-green-600/20 text-green-400 border border-green-600/50 hover:bg-green-600/30'
          : 'bg-red-600/20 text-red-400 border border-red-600/50 hover:bg-red-600/30',
        // Active/holding styles
        isHolding && (isApprove ? 'ring-2 ring-green-500' : 'ring-2 ring-red-500'),
        className
      )}
    >
      {/* Fill animation background */}
      <div
        className={cn(
          'absolute inset-0 transition-all duration-75 origin-left',
          isApprove ? 'bg-green-500' : 'bg-red-500'
        )}
        style={{
          transform: `scaleX(${progress / 100})`,
          opacity: isHolding ? 0.8 : 0,
        }}
      />

      {/* Content */}
      <span className="relative z-10 flex items-center gap-1.5">
        {isExecuting ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : isApprove ? (
          <CheckCircle2 className="h-4 w-4" />
        ) : (
          <XCircle className="h-4 w-4" />
        )}
        <span className={cn(isHolding && 'text-white')}>
          {isExecuting
            ? 'Processing...'
            : isHolding
            ? `Hold ${Math.ceil((holdDuration - (progress / 100) * holdDuration) / 100) / 10}s`
            : children || (isApprove ? 'Approve' : 'Decline')}
        </span>
      </span>

      {/* Progress indicator text */}
      {isHolding && !isExecuting && (
        <span className="absolute inset-0 flex items-center justify-center z-20">
          <span className="text-white font-semibold text-xs opacity-90">
            {Math.round(progress)}%
          </span>
        </span>
      )}
    </button>
  )
}
