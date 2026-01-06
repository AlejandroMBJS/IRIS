"use client"

import { useState, useRef, useCallback } from "react"
import { Upload, X, File, AlertCircle, Check, Loader2 } from "lucide-react"
import { Button } from "./button"
import { cn } from "@/lib/utils"

// Allowed file types (matching backend)
const ALLOWED_TYPES = [
  "application/pdf",
  "image/jpeg",
  "image/jpg",
  "image/png",
  "image/gif",
  "application/msword",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "text/plain",
]

const ALLOWED_EXTENSIONS = [".pdf", ".jpg", ".jpeg", ".png", ".gif", ".doc", ".docx", ".txt"]

const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB

interface FileUploadProps {
  onUpload: (file: File) => Promise<void>
  onRemove?: () => void
  accept?: string
  maxSize?: number
  disabled?: boolean
  className?: string
  uploadedFileName?: string
  label?: string
}

export function FileUpload({
  onUpload,
  onRemove,
  accept = ALLOWED_EXTENSIONS.join(","),
  maxSize = MAX_FILE_SIZE,
  disabled = false,
  className,
  uploadedFileName,
  label = "Cargar Evidencia",
}: FileUploadProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [uploadedFile, setUploadedFile] = useState<string | null>(uploadedFileName || null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const validateFile = useCallback(
    (file: File): string | null => {
      // Check file size
      if (file.size > maxSize) {
        return `El archivo excede el tamano maximo permitido (${Math.round(maxSize / 1024 / 1024)}MB)`
      }

      // Check file type
      const extension = "." + file.name.split(".").pop()?.toLowerCase()
      if (!ALLOWED_EXTENSIONS.includes(extension)) {
        return `Tipo de archivo no permitido. Archivos permitidos: ${ALLOWED_EXTENSIONS.join(", ")}`
      }

      // Check MIME type
      if (!ALLOWED_TYPES.includes(file.type) && file.type !== "") {
        return "Tipo de archivo no permitido"
      }

      return null
    },
    [maxSize]
  )

  const handleFile = useCallback(
    async (file: File) => {
      setError(null)

      const validationError = validateFile(file)
      if (validationError) {
        setError(validationError)
        return
      }

      setUploading(true)
      try {
        await onUpload(file)
        setUploadedFile(file.name)
      } catch (err: unknown) {
        const errorMessage = err instanceof Error ? err.message : "Error al cargar el archivo"
        setError(errorMessage)
      } finally {
        setUploading(false)
      }
    },
    [onUpload, validateFile]
  )

  const handleDrop = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault()
      setIsDragging(false)

      if (disabled || uploading) return

      const files = e.dataTransfer.files
      if (files.length > 0) {
        handleFile(files[0])
      }
    },
    [disabled, uploading, handleFile]
  )

  const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = e.target.files
      if (files && files.length > 0) {
        handleFile(files[0])
      }
      // Reset input to allow uploading same file again
      if (fileInputRef.current) {
        fileInputRef.current.value = ""
      }
    },
    [handleFile]
  )

  const handleRemove = useCallback(() => {
    setUploadedFile(null)
    setError(null)
    if (onRemove) {
      onRemove()
    }
  }, [onRemove])

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return bytes + " B"
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB"
    return (bytes / 1024 / 1024).toFixed(1) + " MB"
  }

  return (
    <div className={cn("space-y-2", className)}>
      {label && <label className="text-sm font-medium text-slate-300">{label}</label>}

      {uploadedFile ? (
        <div className="flex items-center justify-between p-3 bg-green-500/10 border border-green-500/30 rounded-lg">
          <div className="flex items-center gap-2">
            <Check className="h-4 w-4 text-green-400" />
            <File className="h-4 w-4 text-slate-400" />
            <span className="text-sm text-slate-300 truncate max-w-[200px]">{uploadedFile}</span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRemove}
            className="h-6 w-6 p-0 text-slate-400 hover:text-red-400"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      ) : (
        <div
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          className={cn(
            "relative border-2 border-dashed rounded-lg p-4 transition-colors cursor-pointer",
            isDragging
              ? "border-blue-500 bg-blue-500/10"
              : "border-slate-600 hover:border-slate-500",
            disabled && "opacity-50 cursor-not-allowed"
          )}
          onClick={() => !disabled && !uploading && fileInputRef.current?.click()}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept={accept}
            onChange={handleFileInput}
            disabled={disabled || uploading}
            className="hidden"
          />

          <div className="flex flex-col items-center gap-2 text-center">
            {uploading ? (
              <>
                <Loader2 className="h-8 w-8 text-blue-400 animate-spin" />
                <span className="text-sm text-slate-400">Subiendo archivo...</span>
              </>
            ) : (
              <>
                <Upload className="h-8 w-8 text-slate-400" />
                <div className="text-sm">
                  <span className="text-blue-400 font-medium">Click para seleccionar</span>
                  <span className="text-slate-400"> o arrastra aqui</span>
                </div>
                <span className="text-xs text-slate-500">
                  PDF, JPG, PNG, GIF, DOC, DOCX, TXT (max {formatFileSize(maxSize)})
                </span>
              </>
            )}
          </div>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 p-2 bg-red-500/10 border border-red-500/30 rounded-lg">
          <AlertCircle className="h-4 w-4 text-red-400" />
          <span className="text-sm text-red-400">{error}</span>
        </div>
      )}
    </div>
  )
}
