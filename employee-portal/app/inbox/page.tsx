"use client"

import { useEffect, useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import {
  Inbox,
  Send,
  Mail,
  MailOpen,
  Clock,
  User,
  Plus,
  ChevronRight,
  ArrowLeft,
  Reply,
  Archive,
  Trash2,
  Search,
  RefreshCcw,
  MessageSquare,
  Bell,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { PortalLayout } from "@/components/layout/portal-layout"
import { isAuthenticated, getCurrentUser } from "@/lib/auth"
import {
  messageApi,
  Message,
  MessagesResponse,
} from "@/lib/api-client"
import { useToast } from "@/hooks/use-toast"

type TabType = 'inbox' | 'sent';

export default function InboxPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [user, setUser] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<TabType>('inbox')
  const [messages, setMessages] = useState<Message[]>([])
  const [totalMessages, setTotalMessages] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [unreadCount, setUnreadCount] = useState(0)
  const [statusFilter, setStatusFilter] = useState('all')

  // Selected message view
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null)
  const [messageLoading, setMessageLoading] = useState(false)

  // Compose dialog
  const [composeOpen, setComposeOpen] = useState(false)
  const [composing, setComposing] = useState(false)
  const [recipients, setRecipients] = useState<{ id: string; full_name: string; email: string; role: string }[]>([])
  const [recipientSearch, setRecipientSearch] = useState('')
  const [selectedRecipient, setSelectedRecipient] = useState('')
  const [subject, setSubject] = useState('')
  const [body, setBody] = useState('')

  // Reply state
  const [replyText, setReplyText] = useState('')
  const [replying, setReplying] = useState(false)

  // Load messages based on active tab
  const loadMessages = useCallback(async () => {
    setLoading(true)
    try {
      let response: MessagesResponse
      if (activeTab === 'inbox') {
        response = await messageApi.getInbox(currentPage, 20, statusFilter)
      } else {
        response = await messageApi.getSent(currentPage, 20)
      }
      setMessages(response.messages || [])
      setTotalMessages(response.total)
      setTotalPages(response.total_pages)
    } catch (error) {
      console.error('Error loading messages:', error)
      toast({
        title: "Error",
        description: "Could not load messages",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }, [activeTab, currentPage, statusFilter, toast])

  // Load unread count
  const loadUnreadCount = useCallback(async () => {
    try {
      const response = await messageApi.getUnreadCount()
      setUnreadCount(response.unread_count)
    } catch (error) {
      console.error('Error loading unread count:', error)
    }
  }, [])

  // Load recipients for compose
  const loadRecipients = useCallback(async (search: string) => {
    try {
      const response = await messageApi.getRecipients(search)
      setRecipients(response.users || [])
    } catch (error) {
      console.error('Error loading recipients:', error)
    }
  }, [])

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/auth/login")
      return
    }
    setUser(getCurrentUser())
    loadMessages()
    loadUnreadCount()
  }, [router, loadMessages, loadUnreadCount])

  // Reload messages when tab changes
  useEffect(() => {
    setCurrentPage(1)
    setSelectedMessage(null)
  }, [activeTab])

  // Search recipients when typing
  useEffect(() => {
    const timer = setTimeout(() => {
      if (composeOpen) {
        loadRecipients(recipientSearch)
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [recipientSearch, composeOpen, loadRecipients])

  // View message details
  const handleViewMessage = async (message: Message) => {
    setMessageLoading(true)
    setSelectedMessage(message)
    try {
      const fullMessage = await messageApi.get(message.id)
      setSelectedMessage(fullMessage)
      // Reload inbox to update read status
      if (activeTab === 'inbox' && message.status === 'unread') {
        loadUnreadCount()
        loadMessages()
      }
    } catch (error) {
      console.error('Error loading message:', error)
    } finally {
      setMessageLoading(false)
    }
  }

  // Send new message
  const handleSendMessage = async () => {
    if (!selectedRecipient || !subject.trim() || !body.trim()) {
      toast({
        title: "Error",
        description: "Please fill in all fields",
        variant: "destructive",
      })
      return
    }

    setComposing(true)
    try {
      await messageApi.send({
        recipient_id: selectedRecipient,
        subject: subject.trim(),
        body: body.trim(),
      })
      toast({
        title: "Message Sent",
        description: "Your message has been sent successfully",
      })
      setComposeOpen(false)
      setSelectedRecipient('')
      setSubject('')
      setBody('')
      if (activeTab === 'sent') {
        loadMessages()
      }
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not send message",
        variant: "destructive",
      })
    } finally {
      setComposing(false)
    }
  }

  // Reply to message
  const handleReply = async () => {
    if (!selectedMessage || !replyText.trim()) {
      return
    }

    setReplying(true)
    try {
      await messageApi.reply(selectedMessage.id, { body: replyText.trim() })
      toast({
        title: "Reply Sent",
        description: "Your reply has been sent",
      })
      setReplyText('')
      // Reload the message to see the reply
      const updated = await messageApi.get(selectedMessage.id)
      setSelectedMessage(updated)
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not send reply",
        variant: "destructive",
      })
    } finally {
      setReplying(false)
    }
  }

  // Archive message
  const handleArchive = async (messageId: string) => {
    try {
      await messageApi.archive(messageId)
      toast({
        title: "Message Archived",
        description: "The message has been archived",
      })
      setSelectedMessage(null)
      loadMessages()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not archive message",
        variant: "destructive",
      })
    }
  }

  // Delete message
  const handleDelete = async (messageId: string) => {
    try {
      await messageApi.delete(messageId)
      toast({
        title: "Message Deleted",
        description: "The message has been deleted",
      })
      setSelectedMessage(null)
      loadMessages()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.message || "Could not delete message",
        variant: "destructive",
      })
    }
  }

  // Format date
  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffHours = diffMs / (1000 * 60 * 60)
    const diffDays = diffMs / (1000 * 60 * 60 * 24)

    if (diffHours < 1) {
      return "A few minutes ago"
    } else if (diffHours < 24) {
      return `${Math.floor(diffHours)}h ago`
    } else if (diffDays < 7) {
      return `${Math.floor(diffDays)}d ago`
    } else {
      return date.toLocaleDateString("en-US", {
        day: "2-digit",
        month: "short",
      })
    }
  }

  // Get role label
  const getRoleLabel = (role: string) => {
    const labels: Record<string, string> = {
      admin: 'Administrator',
      hr: 'Human Resources',
      hr_and_pr: 'HR & Payroll',
      hr_blue_gray: 'HR Operations',
      hr_white: 'HR Administrative',
      supervisor: 'Supervisor',
      manager: 'Manager',
      sup_and_gm: 'Supervisor/GM',
      employee: 'Employee',
    }
    return labels[role] || role
  }

  return (
    <PortalLayout>
      <div className="flex h-[calc(100vh-120px)] gap-6">
        {/* Sidebar */}
        <div className="w-64 flex-shrink-0">
          <div className="bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 p-4 h-full">
            {/* Compose Button */}
            <Button
              onClick={() => setComposeOpen(true)}
              className="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 mb-6"
            >
              <Plus size={18} className="mr-2" />
              New Message
            </Button>

            {/* Navigation */}
            <nav className="space-y-2">
              <button
                onClick={() => setActiveTab('inbox')}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                  activeTab === 'inbox'
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                    : 'text-slate-400 hover:bg-slate-700/50 hover:text-white'
                }`}
              >
                <Inbox size={20} />
                <span className="flex-1 text-left">Inbox</span>
                {unreadCount > 0 && (
                  <Badge className="bg-blue-500 text-white">{unreadCount}</Badge>
                )}
              </button>

              <button
                onClick={() => setActiveTab('sent')}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                  activeTab === 'sent'
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                    : 'text-slate-400 hover:bg-slate-700/50 hover:text-white'
                }`}
              >
                <Send size={20} />
                <span className="flex-1 text-left">Sent</span>
              </button>
            </nav>

            {/* Stats */}
            <div className="mt-6 pt-6 border-t border-slate-700/50">
              <div className="text-xs text-slate-500 uppercase tracking-wider mb-3">Summary</div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between text-slate-400">
                  <span>Total messages</span>
                  <span className="text-white">{totalMessages}</span>
                </div>
                <div className="flex justify-between text-slate-400">
                  <span>Unread</span>
                  <span className="text-blue-400">{unreadCount}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 flex gap-6">
          {/* Message List */}
          <div className="flex-1 bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden flex flex-col">
            {/* Header */}
            <div className="p-4 border-b border-slate-700/50 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white flex items-center gap-2">
                {activeTab === 'inbox' ? <Inbox size={20} /> : <Send size={20} />}
                {activeTab === 'inbox' ? 'Inbox' : 'Sent'}
              </h2>
              <div className="flex items-center gap-2">
                {activeTab === 'inbox' && (
                  <Select value={statusFilter} onValueChange={setStatusFilter}>
                    <SelectTrigger className="w-32 bg-slate-800 border-slate-600">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All</SelectItem>
                      <SelectItem value="unread">Unread</SelectItem>
                      <SelectItem value="read">Read</SelectItem>
                    </SelectContent>
                  </Select>
                )}
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => loadMessages()}
                  className="text-slate-400 hover:text-white"
                >
                  <RefreshCcw size={18} />
                </Button>
              </div>
            </div>

            {/* Message List */}
            <div className="flex-1 overflow-y-auto">
              {loading ? (
                <div className="flex items-center justify-center h-64">
                  <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
                </div>
              ) : messages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-64 text-slate-400">
                  <Mail size={48} className="mb-4 text-slate-600" />
                  <p>No messages</p>
                </div>
              ) : (
                <div className="divide-y divide-slate-700/50">
                  {messages.map((message) => (
                    <div
                      key={message.id}
                      onClick={() => handleViewMessage(message)}
                      className={`p-4 cursor-pointer transition-all hover:bg-slate-700/30 ${
                        selectedMessage?.id === message.id ? 'bg-slate-700/40' : ''
                      } ${message.status === 'unread' ? 'bg-blue-500/5' : ''}`}
                    >
                      <div className="flex items-start gap-3">
                        <div className={`p-2 rounded-lg ${
                          message.status === 'unread'
                            ? 'bg-blue-500/20 text-blue-400'
                            : 'bg-slate-700/50 text-slate-400'
                        }`}>
                          {message.status === 'unread' ? <Mail size={18} /> : <MailOpen size={18} />}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`font-medium truncate ${
                              message.status === 'unread' ? 'text-white' : 'text-slate-300'
                            }`}>
                              {activeTab === 'inbox'
                                ? message.sender?.full_name || 'Unknown'
                                : message.recipient?.full_name || 'Unknown'}
                            </span>
                            {message.type === 'announcement_question' && (
                              <Badge variant="outline" className="text-xs border-amber-500/30 text-amber-400">
                                <Bell size={10} className="mr-1" />
                                Announcement
                              </Badge>
                            )}
                            {message.type === 'system' && (
                              <Badge variant="outline" className="text-xs border-purple-500/30 text-purple-400">
                                System
                              </Badge>
                            )}
                          </div>
                          <p className={`text-sm truncate ${
                            message.status === 'unread' ? 'text-slate-200' : 'text-slate-400'
                          }`}>
                            {message.subject}
                          </p>
                          <p className="text-xs text-slate-500 truncate mt-1">
                            {message.body.substring(0, 80)}...
                          </p>
                        </div>
                        <div className="text-xs text-slate-500 flex-shrink-0">
                          {formatDate(message.created_at)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="p-4 border-t border-slate-700/50 flex items-center justify-between">
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={currentPage === 1}
                  onClick={() => setCurrentPage(p => p - 1)}
                  className="text-slate-400"
                >
                  Previous
                </Button>
                <span className="text-sm text-slate-400">
                  Page {currentPage} of {totalPages}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={currentPage === totalPages}
                  onClick={() => setCurrentPage(p => p + 1)}
                  className="text-slate-400"
                >
                  Next
                </Button>
              </div>
            )}
          </div>

          {/* Message Detail */}
          {selectedMessage && (
            <div className="w-96 bg-gradient-to-br from-slate-800/60 to-slate-900/60 backdrop-blur-xl rounded-2xl border border-slate-700/50 overflow-hidden flex flex-col">
              {/* Header */}
              <div className="p-4 border-b border-slate-700/50">
                <div className="flex items-center justify-between mb-4">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedMessage(null)}
                    className="text-slate-400 hover:text-white"
                  >
                    <ArrowLeft size={18} className="mr-1" />
                    Close
                  </Button>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleArchive(selectedMessage.id)}
                      className="text-slate-400 hover:text-amber-400"
                      title="Archive"
                    >
                      <Archive size={18} />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(selectedMessage.id)}
                      className="text-slate-400 hover:text-red-400"
                      title="Delete"
                    >
                      <Trash2 size={18} />
                    </Button>
                  </div>
                </div>
                <h3 className="text-lg font-semibold text-white mb-2">
                  {selectedMessage.subject}
                </h3>
                <div className="flex items-center gap-2 text-sm">
                  <User size={14} className="text-slate-500" />
                  <span className="text-slate-300">
                    {activeTab === 'inbox'
                      ? selectedMessage.sender?.full_name
                      : selectedMessage.recipient?.full_name}
                  </span>
                  <Badge variant="outline" className="text-xs border-slate-600 text-slate-400">
                    {getRoleLabel(
                      activeTab === 'inbox'
                        ? selectedMessage.sender?.role || ''
                        : selectedMessage.recipient?.role || ''
                    )}
                  </Badge>
                </div>
                <div className="flex items-center gap-2 text-xs text-slate-500 mt-2">
                  <Clock size={12} />
                  <span>{new Date(selectedMessage.created_at).toLocaleString('en-US')}</span>
                </div>
              </div>

              {/* Message Body */}
              <div className="flex-1 overflow-y-auto p-4">
                {messageLoading ? (
                  <div className="flex items-center justify-center h-32">
                    <div className="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
                  </div>
                ) : (
                  <>
                    {/* Original Message */}
                    <div className="prose prose-invert max-w-none">
                      <p className="text-slate-300 whitespace-pre-wrap">{selectedMessage.body}</p>
                    </div>

                    {/* Linked Announcement */}
                    {selectedMessage.announcement && (
                      <div className="mt-4 p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg">
                        <div className="flex items-center gap-2 text-xs text-amber-400 mb-2">
                          <Bell size={12} />
                          <span>Question about announcement</span>
                        </div>
                        <p className="text-sm font-medium text-white">
                          {selectedMessage.announcement.title}
                        </p>
                      </div>
                    )}

                    {/* Replies */}
                    {selectedMessage.replies && selectedMessage.replies.length > 0 && (
                      <div className="mt-6 space-y-4">
                        <div className="text-xs text-slate-500 uppercase tracking-wider">
                          Replies ({selectedMessage.replies.length})
                        </div>
                        {selectedMessage.replies.map((reply) => (
                          <div
                            key={reply.id}
                            className={`p-3 rounded-lg ${
                              reply.sender_id === user?.id
                                ? 'bg-blue-500/10 border border-blue-500/30 ml-4'
                                : 'bg-slate-700/30 border border-slate-600/30 mr-4'
                            }`}
                          >
                            <div className="flex items-center gap-2 text-xs text-slate-400 mb-2">
                              <User size={12} />
                              <span>{reply.sender?.full_name}</span>
                              <span>-</span>
                              <span>{formatDate(reply.created_at)}</span>
                            </div>
                            <p className="text-sm text-slate-300 whitespace-pre-wrap">
                              {reply.body}
                            </p>
                          </div>
                        ))}
                      </div>
                    )}
                  </>
                )}
              </div>

              {/* Reply Box */}
              {activeTab === 'inbox' && (
                <div className="p-4 border-t border-slate-700/50">
                  <Textarea
                    placeholder="Write your reply..."
                    value={replyText}
                    onChange={(e) => setReplyText(e.target.value)}
                    className="bg-slate-800 border-slate-600 text-white mb-3 resize-none"
                    rows={3}
                  />
                  <Button
                    onClick={handleReply}
                    disabled={replying || !replyText.trim()}
                    className="w-full bg-blue-600 hover:bg-blue-700"
                  >
                    {replying ? (
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                    ) : (
                      <Reply size={16} className="mr-2" />
                    )}
                    Reply
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Compose Dialog */}
      <Dialog open={composeOpen} onOpenChange={setComposeOpen}>
        <DialogContent className="bg-slate-900 border-slate-700 text-white max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <MessageSquare size={20} className="text-blue-400" />
              New Message
            </DialogTitle>
            <DialogDescription className="text-slate-400">
              Send a message to HR, supervisor, or manager
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {/* Recipient Search */}
            <div className="space-y-2">
              <label className="text-sm text-slate-400">To:</label>
              <Input
                placeholder="Search recipient..."
                value={recipientSearch}
                onChange={(e) => setRecipientSearch(e.target.value)}
                className="bg-slate-800 border-slate-600 text-white"
              />
              {recipients.length > 0 && recipientSearch && (
                <div className="bg-slate-800 border border-slate-600 rounded-lg max-h-40 overflow-y-auto">
                  {recipients.map((recipient) => (
                    <button
                      key={recipient.id}
                      onClick={() => {
                        setSelectedRecipient(recipient.id)
                        setRecipientSearch(recipient.full_name)
                        setRecipients([])
                      }}
                      className="w-full p-3 text-left hover:bg-slate-700/50 flex items-center gap-3"
                    >
                      <div className="w-8 h-8 rounded-full bg-slate-700 flex items-center justify-center">
                        <User size={14} className="text-slate-400" />
                      </div>
                      <div>
                        <p className="text-sm text-white">{recipient.full_name}</p>
                        <p className="text-xs text-slate-500">{getRoleLabel(recipient.role)}</p>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Subject */}
            <div className="space-y-2">
              <label className="text-sm text-slate-400">Subject:</label>
              <Input
                placeholder="Message subject"
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                className="bg-slate-800 border-slate-600 text-white"
              />
            </div>

            {/* Body */}
            <div className="space-y-2">
              <label className="text-sm text-slate-400">Message:</label>
              <Textarea
                placeholder="Write your message..."
                value={body}
                onChange={(e) => setBody(e.target.value)}
                className="bg-slate-800 border-slate-600 text-white min-h-32"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="ghost"
              onClick={() => setComposeOpen(false)}
              className="text-slate-400"
            >
              Cancel
            </Button>
            <Button
              onClick={handleSendMessage}
              disabled={composing || !selectedRecipient || !subject.trim() || !body.trim()}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {composing ? (
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
              ) : (
                <Send size={16} className="mr-2" />
              )}
              Send
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PortalLayout>
  )
}
