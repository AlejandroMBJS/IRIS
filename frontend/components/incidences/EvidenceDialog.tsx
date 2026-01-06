"use client"

import { Paperclip, FileText, Download, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { FileUpload } from "@/components/ui/file-upload"
import { Incidence, IncidenceEvidence } from "@/lib/api-client"
import { formatFileSize } from "./constants"

interface EvidenceDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  incidence: Incidence | null
  evidences: IncidenceEvidence[]
  loading: boolean
  onUpload: (file: File) => Promise<void>
  onDownload: (evidenceId: string, fileName: string) => void
  onDelete: (evidenceId: string) => void
}

export function EvidenceDialog({
  isOpen,
  onOpenChange,
  incidence,
  evidences,
  loading,
  onUpload,
  onDownload,
  onDelete,
}: EvidenceDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Paperclip className="h-5 w-5" />
            Incidence Evidence
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            {incidence && (
              <>
                {incidence.employee?.first_name} {incidence.employee?.last_name} - {incidence.incidence_type?.name}
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Upload Section */}
          <div className="space-y-2">
            <FileUpload
              onUpload={onUpload}
              label="Upload New Evidence"
            />
          </div>

          {/* Evidence List */}
          <div className="space-y-2">
            <Label className="text-slate-300">Attached Files ({evidences.length})</Label>

            {loading ? (
              <div className="text-center text-slate-400 py-4">Loading files...</div>
            ) : evidences.length === 0 ? (
              <div className="text-center text-slate-500 py-4 bg-slate-800/50 rounded-lg border border-slate-700">
                No evidence attached
              </div>
            ) : (
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {evidences.map((evidence) => (
                  <div
                    key={evidence.id}
                    className="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg border border-slate-700"
                  >
                    <div className="flex items-center gap-3 min-w-0 flex-1">
                      <FileText className="h-5 w-5 text-blue-400 flex-shrink-0" />
                      <div className="min-w-0">
                        <div className="text-sm text-white truncate" title={evidence.original_name}>
                          {evidence.original_name}
                        </div>
                        <div className="text-xs text-slate-500">
                          {formatFileSize(evidence.file_size)} - {new Date(evidence.created_at).toLocaleDateString("es-MX")}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDownload(evidence.id, evidence.original_name)}
                        className="text-blue-400 hover:text-blue-300"
                        title="Download"
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(evidence.id)}
                        className="text-red-400 hover:text-red-300"
                        title="Delete"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="border-slate-600"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
