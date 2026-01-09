'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  absenceRequestApi,
  authApi,
  type AbsenceRequest,
  type ApprovalStage,
  type ApprovalAction,
} from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CheckCircle2, XCircle, Clock, RefreshCw, User, Calendar } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { DashboardLayout } from '@/components/layout/dashboard-layout'
import { HoldToConfirmButton } from '@/components/ui/hold-to-confirm-button'

type UserRole = 'supervisor' | 'manager' | 'hr_blue_gray' | 'hr_white' | 'admin' | 'supandgm'

export default function ApprovalsPage() {
  const router = useRouter()
  const { toast } = useToast()

  const [userRole, setUserRole] = useState<UserRole | null>(null)
  const [activeTab, setActiveTab] = useState<ApprovalStage>('SUPERVISOR')
  const [requests, setRequests] = useState<AbsenceRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [actionDialogOpen, setActionDialogOpen] = useState(false)
  const [selectedRequest, setSelectedRequest] = useState<AbsenceRequest | null>(null)
  const [actionType, setActionType] = useState<ApprovalAction>('APPROVED')
  const [comments, setComments] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [pendingCounts, setPendingCounts] = useState({
    supervisor_count: 0,
    general_manager_count: 0,
    hr_blue_gray_count: 0,
    hr_white_count: 0,
  })

  // Auto-refresh every 30 seconds
  useEffect(() => {
    const interval = setInterval(() => {
      loadRequests()
      loadPendingCounts()
    }, 30000)

    return () => clearInterval(interval)
  }, [activeTab])

  // Load user profile and determine role
  useEffect(() => {
    const loadProfile = async () => {
      try {
        const profile = await authApi.getProfile()
        setUserRole(profile.role as UserRole)

        // Set initial tab based on role
        if (profile.role === 'supervisor' || profile.role === 'supandgm') {
          setActiveTab('SUPERVISOR')
        } else if (profile.role === 'manager' || profile.role === 'supandgm') {
          setActiveTab('GENERAL_MANAGER')
        } else if (profile.role === 'hr_blue_gray') {
          setActiveTab('HR_BLUE_GRAY')
        } else if (profile.role === 'hr' || profile.role === 'hr_white') {
          setActiveTab('HR')
        }
      } catch (error) {
        console.error('Failed to load profile:', error)
        router.push('/auth/login')
      }
    }

    loadProfile()
  }, [router])

  // Load requests when tab changes
  useEffect(() => {
    if (userRole) {
      loadRequests()
      loadPendingCounts()
    }
  }, [activeTab, userRole])

  const loadRequests = async () => {
    try {
      setLoading(true)
      const { requests: data } = await absenceRequestApi.getPending(activeTab.toLowerCase())
      setRequests(data)
    } catch (error) {
      console.error('Failed to load requests:', error)
      toast({
        title: 'Error',
        description: 'Failed to load pending requests',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const loadPendingCounts = async () => {
    try {
      const counts = await absenceRequestApi.getCounts()
      setPendingCounts(counts)
    } catch (error) {
      console.error('Failed to load counts:', error)
    }
  }

  // Direct approve handler (no dialog needed - uses hold-to-confirm)
  const handleDirectApprove = async (request: AbsenceRequest) => {
    try {
      await absenceRequestApi.approve(request.id, {
        action: 'APPROVED',
        stage: activeTab,
        comments: '',
      })

      toast({
        title: 'Success',
        description: 'Request approved successfully',
      })

      loadRequests()
      loadPendingCounts()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to approve request',
        variant: 'destructive',
      })
    }
  }

  // Decline still needs dialog for comments
  const handleDeclineClick = (request: AbsenceRequest) => {
    setSelectedRequest(request)
    setActionType('DECLINED')
    setComments('')
    setActionDialogOpen(true)
  }

  const handleSubmitAction = async () => {
    if (!selectedRequest) return

    try {
      setSubmitting(true)

      await absenceRequestApi.approve(selectedRequest.id, {
        action: actionType,
        stage: activeTab,
        comments,
      })

      toast({
        title: 'Success',
        description: `Request ${actionType.toLowerCase()} successfully`,
      })

      setActionDialogOpen(false)
      setSelectedRequest(null)
      setComments('')
      loadRequests()
      loadPendingCounts()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to process request',
        variant: 'destructive',
      })
    } finally {
      setSubmitting(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  const getRequestTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      PAID_LEAVE: 'Paid Leave',
      UNPAID_LEAVE: 'Unpaid Leave',
      VACATION: 'Vacation',
      LATE_ENTRY: 'Late Entry',
      EARLY_EXIT: 'Early Exit',
      SHIFT_CHANGE: 'Shift Change',
      TIME_FOR_TIME: 'Time for Time',
      SICK_LEAVE: 'Sick Leave',
      PERSONAL: 'Personal',
      OTHER: 'Other',
    }
    return labels[type] || type
  }

  const canViewTab = (stage: ApprovalStage): boolean => {
    if (!userRole) return false

    switch (stage) {
      case 'SUPERVISOR':
        return ['supervisor', 'supandgm', 'admin'].includes(userRole)
      case 'GENERAL_MANAGER':
        return ['manager', 'supandgm', 'admin'].includes(userRole)
      case 'HR_BLUE_GRAY':
        return ['hr_blue_gray', 'hr', 'admin'].includes(userRole)
      case 'HR':
        return ['hr', 'hr_white', 'admin'].includes(userRole)
      default:
        return false
    }
  }

  if (!userRole) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-96">
          <div className="text-center">
            <RefreshCw className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-500" />
            <p className="text-slate-400">Loading...</p>
          </div>
        </div>
      </DashboardLayout>
    )
  }

  return (
    <DashboardLayout>
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Absence Request Approvals</h1>
          <p className="text-slate-400 mt-1">Review and approve pending absence requests</p>
        </div>
        <Button onClick={loadRequests} variant="outline" size="sm">
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as ApprovalStage)}>
        <TabsList className="grid w-full grid-cols-2 md:grid-cols-4 bg-slate-800 gap-1">
          {canViewTab('SUPERVISOR') && (
            <TabsTrigger value="SUPERVISOR" className="relative">
              Supervisor
              {pendingCounts.supervisor_count > 0 && (
                <Badge className="ml-2 bg-amber-500 text-white" variant="secondary">
                  {pendingCounts.supervisor_count}
                </Badge>
              )}
            </TabsTrigger>
          )}
          {canViewTab('GENERAL_MANAGER') && (
            <TabsTrigger value="GENERAL_MANAGER" className="relative">
              General Manager
              {pendingCounts.general_manager_count > 0 && (
                <Badge className="ml-2 bg-amber-500 text-white" variant="secondary">
                  {pendingCounts.general_manager_count}
                </Badge>
              )}
            </TabsTrigger>
          )}
          {canViewTab('HR_BLUE_GRAY') && (
            <TabsTrigger value="HR_BLUE_GRAY" className="relative">
              HR (Blue/Gray)
              {pendingCounts.hr_blue_gray_count > 0 && (
                <Badge className="ml-2 bg-amber-500 text-white" variant="secondary">
                  {pendingCounts.hr_blue_gray_count}
                </Badge>
              )}
            </TabsTrigger>
          )}
          {canViewTab('HR') && (
            <TabsTrigger value="HR" className="relative">
              HR (White)
              {pendingCounts.hr_white_count > 0 && (
                <Badge className="ml-2 bg-amber-500 text-white" variant="secondary">
                  {pendingCounts.hr_white_count}
                </Badge>
              )}
            </TabsTrigger>
          )}
        </TabsList>

        {(['SUPERVISOR', 'GENERAL_MANAGER', 'HR_BLUE_GRAY', 'HR'] as ApprovalStage[]).map((stage) => {
          if (!canViewTab(stage)) return null

          return (
            <TabsContent key={stage} value={stage} className="mt-6">
              <Card className="bg-slate-800 border-slate-700">
                <CardHeader>
                  <CardTitle className="text-white flex items-center gap-2">
                    <Clock className="h-5 w-5 text-amber-400" />
                    Pending Requests - {stage.replace('_', ' ')}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <div className="flex items-center justify-center py-12">
                      <RefreshCw className="h-8 w-8 animate-spin text-blue-500" />
                    </div>
                  ) : requests.length === 0 ? (
                    <div className="text-center py-12">
                      <CheckCircle2 className="h-12 w-12 text-green-400 mx-auto mb-4" />
                      <p className="text-slate-400">No pending requests at this stage</p>
                    </div>
                  ) : (
                    <div className="overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow className="border-slate-700">
                          <TableHead className="text-slate-300">Employee</TableHead>
                          <TableHead className="text-slate-300">Type</TableHead>
                          <TableHead className="text-slate-300 hidden md:table-cell">Period</TableHead>
                          <TableHead className="text-slate-300 hidden lg:table-cell">Days</TableHead>
                          <TableHead className="text-slate-300 hidden lg:table-cell">Reason</TableHead>
                          <TableHead className="text-slate-300 hidden md:table-cell">Collar Type</TableHead>
                          <TableHead className="text-slate-300 hidden xl:table-cell">Submitted</TableHead>
                          <TableHead className="text-slate-300 text-right">Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {requests.map((request) => (
                          <TableRow key={request.id} className="border-slate-700">
                            <TableCell>
                              <div className="flex items-center gap-2">
                                <User className="h-4 w-4 text-slate-400" />
                                <div>
                                  <p className="font-medium text-white">
                                    {request.employee?.full_name || 'Unknown'}
                                  </p>
                                  <p className="text-sm text-slate-400">
                                    #{request.employee?.employee_number}
                                  </p>
                                </div>
                              </div>
                            </TableCell>
                            <TableCell>
                              <Badge variant="outline" className="border-blue-500 text-blue-400">
                                {getRequestTypeLabel(request.request_type)}
                              </Badge>
                            </TableCell>
                            <TableCell className="hidden md:table-cell">
                              <div className="flex items-center gap-1 text-slate-300">
                                <Calendar className="h-3 w-3" />
                                <span className="text-sm">
                                  {formatDate(request.start_date)}
                                  {request.start_date !== request.end_date &&
                                    ` - ${formatDate(request.end_date)}`}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell className="text-slate-300 hidden lg:table-cell">{request.total_days}</TableCell>
                            <TableCell className="max-w-xs hidden lg:table-cell">
                              <p className="text-sm text-slate-300 truncate" title={request.reason}>
                                {request.reason}
                              </p>
                            </TableCell>
                            <TableCell className="hidden md:table-cell">
                              <Badge
                                variant="secondary"
                                className={
                                  request.employee?.collar_type === 'white_collar'
                                    ? 'bg-purple-500/20 text-purple-400'
                                    : request.employee?.collar_type === 'blue_collar'
                                    ? 'bg-blue-500/20 text-blue-400'
                                    : 'bg-gray-500/20 text-gray-400'
                                }
                              >
                                {request.employee?.collar_type?.replace('_', ' ') || 'N/A'}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-slate-400 text-sm hidden xl:table-cell">
                              {formatDate(request.created_at)}
                            </TableCell>
                            <TableCell className="text-right">
                              <div className="flex flex-col sm:flex-row gap-2 justify-end items-center">
                                <HoldToConfirmButton
                                  variant="approve"
                                  onConfirm={() => handleDirectApprove(request)}
                                  holdDuration={1000}
                                  className="text-xs sm:text-sm"
                                >
                                  <span className="hidden sm:inline">Hold to </span>Approve
                                </HoldToConfirmButton>
                                <HoldToConfirmButton
                                  variant="decline"
                                  onConfirm={() => handleDeclineClick(request)}
                                  holdDuration={1000}
                                  className="text-xs sm:text-sm"
                                >
                                  <span className="hidden sm:inline">Hold to </span>Decline
                                </HoldToConfirmButton>
                              </div>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          )
        })}
      </Tabs>

      {/* Decline Dialog - requires comments */}
      <Dialog open={actionDialogOpen} onOpenChange={setActionDialogOpen}>
        <DialogContent className="bg-slate-800 border-slate-700 text-white">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-red-400">
              <XCircle className="h-5 w-5" />
              Decline Request
            </DialogTitle>
          </DialogHeader>

          {selectedRequest && (
            <div className="space-y-4">
              <div className="bg-slate-700 p-4 rounded-lg space-y-2">
                <p className="text-sm text-slate-400">Employee</p>
                <p className="font-medium">{selectedRequest.employee?.full_name}</p>

                <p className="text-sm text-slate-400 mt-3">Request Type</p>
                <p className="font-medium">{getRequestTypeLabel(selectedRequest.request_type)}</p>

                <p className="text-sm text-slate-400 mt-3">Period</p>
                <p className="font-medium">
                  {formatDate(selectedRequest.start_date)} -{' '}
                  {formatDate(selectedRequest.end_date)} ({selectedRequest.total_days} days)
                </p>

                <p className="text-sm text-slate-400 mt-3">Reason</p>
                <p className="font-medium">{selectedRequest.reason}</p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  Reason for declining <span className="text-red-400">*</span>
                </label>
                <Textarea
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  placeholder="Please provide a reason for declining this request"
                  className="bg-slate-700 border-slate-600 text-white"
                  rows={4}
                />
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setActionDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleSubmitAction}
              disabled={submitting || !comments.trim()}
              className="bg-red-600 hover:bg-red-700"
            >
              {submitting ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  Confirm Decline
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
    </DashboardLayout>
  )
}
