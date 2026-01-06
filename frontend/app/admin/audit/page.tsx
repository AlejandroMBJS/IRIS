"use client"

import { useEffect, useState } from "react"
import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { auditApi, AuditLog, LoginSession, AuditStatsResponse } from "@/lib/api-client"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ShieldCheck, Search, AlertTriangle, CheckCircle, XCircle, Clock, Users, Activity } from "lucide-react"
import { toast } from "sonner"

export default function AuditPage() {
  const [loginAttempts, setLoginAttempts] = useState<AuditLog[]>([])
  const [loginHistory, setLoginHistory] = useState<LoginSession[]>([])
  const [stats, setStats] = useState<AuditStatsResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [searchEmail, setSearchEmail] = useState("")
  const [eventTypeFilter, setEventTypeFilter] = useState<string>("all")
  const [successFilter, setSuccessFilter] = useState<string>("all")
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const pageSize = 20

  useEffect(() => {
    loadData()
  }, [currentPage, eventTypeFilter, successFilter])

  const loadData = async () => {
    setLoading(true)
    try {
      // Load login attempts
      const attemptsRes = await auditApi.getLoginAttempts(currentPage, pageSize)
      setLoginAttempts(attemptsRes.data)
      setTotalPages(attemptsRes.pagination.total_pages)

      // Load login history
      const historyRes = await auditApi.getLoginHistory(undefined, 1, 50)
      setLoginHistory(historyRes.data)

      // Load stats
      const statsRes = await auditApi.getStats(30)
      setStats(statsRes)
    } catch (error: any) {
      console.error("Failed to load audit data:", error)
      toast.error("Failed to load audit data", {
        description: error.message || "An error occurred while loading audit logs",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = async () => {
    if (!searchEmail.trim()) {
      loadData()
      return
    }

    setLoading(true)
    try {
      const filters: any = {
        page: currentPage,
        page_size: pageSize,
        email: searchEmail,
      }

      if (eventTypeFilter !== "all") {
        filters.event_type = eventTypeFilter
      }

      if (successFilter !== "all") {
        filters.success = successFilter === "success"
      }

      const attemptsRes = await auditApi.getLogs(filters)
      setLoginAttempts(attemptsRes.data)
      setTotalPages(attemptsRes.pagination.total_pages)
    } catch (error: any) {
      console.error("Failed to search audit logs:", error)
      toast.error("Search failed", {
        description: error.message,
      })
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    })
  }

  const formatDuration = (seconds?: number) => {
    if (!seconds) return "N/A"
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    }
    return `${minutes}m`
  }

  const getEventBadge = (eventType: string, success: boolean) => {
    if (eventType === "login_success" || (eventType === "login_attempt" && success)) {
      return <Badge className="bg-green-600"><CheckCircle className="w-3 h-3 mr-1" /> Success</Badge>
    } else if (eventType === "login_failure" || (eventType === "login_attempt" && !success)) {
      return <Badge className="bg-red-600"><XCircle className="w-3 h-3 mr-1" /> Failed</Badge>
    } else if (eventType === "token_refresh") {
      return <Badge className="bg-blue-600"><Clock className="w-3 h-3 mr-1" /> Token Refresh</Badge>
    } else if (eventType === "logout") {
      return <Badge className="bg-gray-600"><AlertTriangle className="w-3 h-3 mr-1" /> Logout</Badge>
    } else if (eventType === "password_change") {
      return <Badge className="bg-purple-600"><ShieldCheck className="w-3 h-3 mr-1" /> Password Change</Badge>
    }
    return <Badge>{eventType}</Badge>
  }

  return (
    <DashboardLayout>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
              <ShieldCheck className="w-8 h-8 text-blue-400" />
              Audit Logs
            </h1>
            <p className="text-slate-400 mt-1">
              Detailed system activity and security monitoring
            </p>
          </div>
        </div>

        {/* Statistics Cards */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="pb-2">
                <CardDescription className="text-slate-400">Total Attempts</CardDescription>
                <CardTitle className="text-2xl text-white">{stats.total_attempts}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-slate-500">Last 30 days</p>
              </CardContent>
            </Card>

            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="pb-2">
                <CardDescription className="text-slate-400">Successful Logins</CardDescription>
                <CardTitle className="text-2xl text-green-400">{stats.successful_attempts}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-slate-500">
                  {stats.success_rate.toFixed(1)}% success rate
                </p>
              </CardContent>
            </Card>

            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="pb-2">
                <CardDescription className="text-slate-400">Failed Attempts</CardDescription>
                <CardTitle className="text-2xl text-red-400">{stats.failed_attempts}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-slate-500">Security monitoring</p>
              </CardContent>
            </Card>

            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="pb-2">
                <CardDescription className="text-slate-400">Unique Users</CardDescription>
                <CardTitle className="text-2xl text-blue-400">{stats.unique_users}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-xs text-slate-500">Active users</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Tabs */}
        <Tabs defaultValue="login-attempts" className="w-full">
          <TabsList className="bg-slate-800/50">
            <TabsTrigger value="login-attempts" className="data-[state=active]:bg-blue-600">
              <AlertTriangle className="w-4 h-4 mr-2" />
              Login Attempts
            </TabsTrigger>
            <TabsTrigger value="login-history" className="data-[state=active]:bg-blue-600">
              <Clock className="w-4 h-4 mr-2" />
              Login History
            </TabsTrigger>
          </TabsList>

          {/* Login Attempts Tab */}
          <TabsContent value="login-attempts" className="space-y-4">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <div className="flex flex-col md:flex-row gap-4">
                  <div className="flex-1">
                    <div className="flex gap-2">
                      <Input
                        placeholder="Search by email..."
                        value={searchEmail}
                        onChange={(e) => setSearchEmail(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                        className="bg-slate-900/50 border-slate-700 text-white"
                      />
                      <Button onClick={handleSearch} className="bg-blue-600 hover:bg-blue-700">
                        <Search className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>

                  <Select value={eventTypeFilter} onValueChange={setEventTypeFilter}>
                    <SelectTrigger className="w-[200px] bg-slate-900/50 border-slate-700 text-white">
                      <SelectValue placeholder="Event Type" />
                    </SelectTrigger>
                    <SelectContent className="bg-slate-800 border-slate-700">
                      <SelectItem value="all">All Events</SelectItem>
                      <SelectItem value="login_attempt">Login Attempt</SelectItem>
                      <SelectItem value="login_success">Login Success</SelectItem>
                      <SelectItem value="login_failure">Login Failure</SelectItem>
                      <SelectItem value="token_refresh">Token Refresh</SelectItem>
                      <SelectItem value="logout">Logout</SelectItem>
                    </SelectContent>
                  </Select>

                  <Select value={successFilter} onValueChange={setSuccessFilter}>
                    <SelectTrigger className="w-[150px] bg-slate-900/50 border-slate-700 text-white">
                      <SelectValue placeholder="Status" />
                    </SelectTrigger>
                    <SelectContent className="bg-slate-800 border-slate-700">
                      <SelectItem value="all">All Status</SelectItem>
                      <SelectItem value="success">Success</SelectItem>
                      <SelectItem value="failure">Failure</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="border-slate-700">
                        <TableHead className="text-slate-300">Date/Time</TableHead>
                        <TableHead className="text-slate-300">Email</TableHead>
                        <TableHead className="text-slate-300">Event</TableHead>
                        <TableHead className="text-slate-300">Status</TableHead>
                        <TableHead className="text-slate-300">IP Address</TableHead>
                        <TableHead className="text-slate-300">Failure Reason</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {loading ? (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center py-8 text-slate-400">
                            Loading audit logs...
                          </TableCell>
                        </TableRow>
                      ) : loginAttempts.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center py-8 text-slate-400">
                            No audit logs found
                          </TableCell>
                        </TableRow>
                      ) : (
                        loginAttempts.map((log) => (
                          <TableRow key={log.id} className="border-slate-700">
                            <TableCell className="text-slate-300 text-sm">
                              {formatDate(log.created_at)}
                            </TableCell>
                            <TableCell className="text-slate-300 font-mono text-sm">
                              {log.email}
                            </TableCell>
                            <TableCell>
                              <Badge variant="outline" className="text-xs">
                                {log.event_type.replace("_", " ").toUpperCase()}
                              </Badge>
                            </TableCell>
                            <TableCell>{getEventBadge(log.event_type, log.success)}</TableCell>
                            <TableCell className="text-slate-400 text-sm font-mono">
                              {log.ip_address || "N/A"}
                            </TableCell>
                            <TableCell className="text-red-400 text-sm">
                              {log.failure_reason || "-"}
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-4">
                    <p className="text-sm text-slate-400">
                      Page {currentPage} of {totalPages}
                    </p>
                    <div className="flex gap-2">
                      <Button
                        onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                        variant="outline"
                        className="bg-slate-900/50 border-slate-700"
                      >
                        Previous
                      </Button>
                      <Button
                        onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                        variant="outline"
                        className="bg-slate-900/50 border-slate-700"
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* Login History Tab */}
          <TabsContent value="login-history" className="space-y-4">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle className="text-white">Login History & Active Sessions</CardTitle>
                <CardDescription className="text-slate-400">
                  Track user login sessions, duration, and activity
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow className="border-slate-700">
                        <TableHead className="text-slate-300">User</TableHead>
                        <TableHead className="text-slate-300">Login Time</TableHead>
                        <TableHead className="text-slate-300">Logout Time</TableHead>
                        <TableHead className="text-slate-300">Duration</TableHead>
                        <TableHead className="text-slate-300">IP Address</TableHead>
                        <TableHead className="text-slate-300">Status</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {loading ? (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center py-8 text-slate-400">
                            Loading login history...
                          </TableCell>
                        </TableRow>
                      ) : loginHistory.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={6} className="text-center py-8 text-slate-400">
                            No login history found
                          </TableCell>
                        </TableRow>
                      ) : (
                        loginHistory.map((session) => (
                          <TableRow key={session.id} className="border-slate-700">
                            <TableCell className="text-slate-300">
                              <div>
                                <p className="font-medium">{session.user?.full_name || session.email}</p>
                                <p className="text-xs text-slate-500 font-mono">{session.email}</p>
                              </div>
                            </TableCell>
                            <TableCell className="text-slate-300 text-sm">
                              {formatDate(session.login_at)}
                            </TableCell>
                            <TableCell className="text-slate-300 text-sm">
                              {session.logout_at ? formatDate(session.logout_at) : "-"}
                            </TableCell>
                            <TableCell className="text-slate-300 text-sm">
                              {formatDuration(
                                session.logout_at
                                  ? Math.floor(
                                      (new Date(session.logout_at).getTime() -
                                        new Date(session.login_at).getTime()) /
                                        1000
                                    )
                                  : Math.floor(
                                      (new Date(session.last_activity).getTime() -
                                        new Date(session.login_at).getTime()) /
                                        1000
                                    )
                              )}
                            </TableCell>
                            <TableCell className="text-slate-400 text-sm font-mono">
                              {session.ip_address || "N/A"}
                            </TableCell>
                            <TableCell>
                              {session.is_active ? (
                                <Badge className="bg-green-600">
                                  <Activity className="w-3 h-3 mr-1" /> Active
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="text-slate-400">
                                  Ended
                                </Badge>
                              )}
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </DashboardLayout>
  )
}
