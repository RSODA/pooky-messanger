<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

type AuthMode = 'login' | 'register'
type ConnectionState = 'demo' | 'connected' | 'offline'
type ChatSource = 'demo' | 'remote'
type Sender = 'me' | 'system' | string
type WsState = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

interface BackendUser {
  username: string
  first_name: string
  bio: string | null
  avatar: string | null
  created_at: string
}

interface AuthResponse {
  token: string
}

interface GetMeResponse {
  user: BackendUser
}

interface ChatMessage {
  id: string
  sender: Sender
  content: string
  createdAt: string
}

interface ChatContact {
  id: string
  title: string
  username: string
  subtitle: string
  status: string
  accent: string
  avatar?: string
  online: boolean
  unread: number
  messages: ChatMessage[]
  source?: ChatSource
  messagesLoaded?: boolean
  messagesLoading?: boolean
}

interface BackendMessage {
  id: string
  conversation_id: string
  sender_id: string
  content: string
  created_at: string
}

interface WsMessagePayload {
  from: string
  content: string
  conversation_id: string
  created_at: string
}

interface WsEvent {
  event_type: 'new_message' | 'user_online' | 'user_offline' | string
  payload?: Partial<WsMessagePayload>
}

interface ApiOptions extends Omit<RequestInit, 'body'> {
  json?: unknown
}

class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

type IconName =
  | 'message'
  | 'users'
  | 'bell'
  | 'settings'
  | 'logout'
  | 'search'
  | 'plus'
  | 'send'
  | 'heart'
  | 'previous'
  | 'next'
  | 'pause'
  | 'music'
  | 'open'
  | 'user'
  | 'image'
  | 'phone'
  | 'more'
  | 'x'

const TOKEN_STORAGE_KEY = 'pooky-msg:token'
const API_BASE = import.meta.env.VITE_API_BASE ?? ''
const CONVERSATIONS_BASE = (import.meta.env.VITE_CONVERSATIONS_BASE ?? '/api/v1/conversations').replace(
  /\/$/,
  '',
)
const WS_PATH = import.meta.env.VITE_WS_PATH ?? '/api/v1/ws'

const iconPaths: Record<IconName, string> = {
  message:
    '<path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z"/>',
  users:
    '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  bell:
    '<path d="M6 8a6 6 0 0 1 12 0c0 7 3 7 3 9H3c0-2 3-2 3-9"/><path d="M10 21a2 2 0 0 0 4 0"/>',
  settings:
    '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 8.92 4a1.65 1.65 0 0 0 1-1.51V2a2 2 0 0 1 4 0v.09A1.65 1.65 0 0 0 15 3.6a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09A1.65 1.65 0 0 0 19.4 15Z"/>',
  logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5"/><path d="M21 12H9"/>',
  search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
  plus: '<path d="M12 5v14"/><path d="M5 12h14"/>',
  send: '<path d="m22 2-7 20-4-9-9-4Z"/><path d="M22 2 11 13"/>',
  heart:
    '<path d="M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1-1.1a5.5 5.5 0 1 0-7.8 7.8l1 1L12 21l7.8-7.6 1-1a5.5 5.5 0 0 0 0-7.8Z"/>',
  previous: '<path d="m19 20-7-8 7-8v16Z"/><path d="M5 19V5"/>',
  next: '<path d="m5 4 7 8-7 8V4Z"/><path d="M19 5v14"/>',
  pause: '<path d="M10 4H6v16h4V4Z"/><path d="M18 4h-4v16h4V4Z"/>',
  music:
    '<circle cx="8" cy="18" r="4"/><path d="M12 18V3l8 2v11"/><circle cx="18" cy="16" r="4"/>',
  open:
    '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
  user: '<path d="M20 21a8 8 0 1 0-16 0"/><circle cx="12" cy="7" r="4"/>',
  image:
    '<rect width="18" height="18" x="3" y="3" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.1-3.1a2 2 0 0 0-2.8 0L6 21"/>',
  phone:
    '<path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.86 19.86 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.86 19.86 0 0 1 2.12 4.2 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.12.9.32 1.78.6 2.63a2 2 0 0 1-.45 2.11L8 9.7a16 16 0 0 0 6 6l1.24-1.26a2 2 0 0 1 2.11-.45c.85.28 1.73.48 2.63.6A2 2 0 0 1 22 16.92Z"/>',
  more: '<circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/>',
  x: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
}

const navItems: Array<{ id: string; icon: IconName; label: string }> = [
  { id: 'inbox', icon: 'message', label: 'Chats' },
  { id: 'people', icon: 'users', label: 'People' },
  { id: 'alerts', icon: 'bell', label: 'Alerts' },
  { id: 'settings', icon: 'settings', label: 'Settings' },
]

const chats = ref<ChatContact[]>([
  {
    id: 'lunar',
    title: 'Lunar',
    username: 'lunar',
    subtitle: 'yo, are we still meeting tomorrow?',
    status: 'online',
    accent: 'linear-gradient(145deg, #151515, #3b3d42 45%, #101010)',
    online: true,
    unread: 2,
    messages: [
      {
        id: 'lunar-1',
        sender: 'lunar',
        content: 'yo, are we still meeting tomorrow?',
        createdAt: '2026-05-16T21:48:00',
      },
      {
        id: 'lunar-2',
        sender: 'me',
        content: 'yeah. i will push the frontend first and then join.',
        createdAt: '2026-05-16T21:50:00',
      },
    ],
  },
  {
    id: 'elysian',
    title: 'elysian',
    username: 'elysian',
    subtitle: 'sent a photo',
    status: 'last seen 5 min ago',
    accent: 'linear-gradient(145deg, #f2f2f2, #90949c 45%, #111)',
    online: false,
    unread: 1,
    messages: [
      {
        id: 'elysian-1',
        sender: 'elysian',
        content: 'sent a photo',
        createdAt: '2026-05-16T21:12:00',
      },
    ],
  },
  {
    id: 'nox',
    title: 'nox',
    username: 'nox',
    subtitle: 'that was crazy',
    status: 'online',
    accent: 'linear-gradient(145deg, #050505, #21242a 40%, #686868)',
    online: true,
    unread: 0,
    messages: [
      {
        id: 'nox-1',
        sender: 'nox',
        content: 'that was crazy',
        createdAt: '2026-05-16T20:57:00',
      },
      {
        id: 'nox-2',
        sender: 'me',
        content: 'i know. the build finally stopped yelling.',
        createdAt: '2026-05-16T20:58:00',
      },
    ],
  },
  {
    id: 'daze',
    title: 'daze',
    username: 'daze',
    subtitle: 'typing...',
    status: 'typing...',
    accent: 'linear-gradient(145deg, #e9e9e9, #bbb, #282828)',
    online: true,
    unread: 0,
    messages: [
      {
        id: 'daze-1',
        sender: 'daze',
        content: 'typing...',
        createdAt: '2026-05-16T20:43:00',
      },
    ],
  },
  {
    id: 'maya',
    title: 'maya',
    username: 'maya',
    subtitle: 'see you soon <3',
    status: 'away',
    accent: 'linear-gradient(145deg, #0e0e0e, #232323 52%, #555)',
    online: false,
    unread: 0,
    messages: [
      {
        id: 'maya-1',
        sender: 'maya',
        content: 'see you soon <3',
        createdAt: '2026-05-16T18:22:00',
      },
    ],
  },
  {
    id: 'system',
    title: 'system',
    username: 'system',
    subtitle: 'System update: v2.1 is now live.',
    status: 'local',
    accent: 'linear-gradient(145deg, #111, #3b3f46, #080808)',
    online: false,
    unread: 0,
    messages: [
      {
        id: 'system-1',
        sender: 'system',
        content: 'System update: v2.1 is now live.',
        createdAt: '2026-05-15T18:10:00',
      },
    ],
  },
])

const demoChatsSnapshot = chats.value.map((chat) => ({
  ...chat,
  source: 'demo' as const,
  messagesLoaded: true,
  messages: chat.messages.map((message) => ({ ...message })),
}))

const token = ref(localStorage.getItem(TOKEN_STORAGE_KEY) ?? '')
const profile = ref<BackendUser>({
  username: 'voidz',
  first_name: 'voidz',
  bio: 'music, code and late nights. building things in the dark.',
  avatar: null,
  created_at: new Date().toISOString(),
})
const connection = ref<ConnectionState>(token.value ? 'offline' : 'demo')
const selectedChatId = ref('lunar')
const searchQuery = ref('')
const draftMessage = ref('')
const newChatUsername = ref('')
const authOpen = ref(false)
const authMode = ref<AuthMode>('login')
const authPending = ref(false)
const authError = ref('')
const profileOpen = ref(false)
const profilePending = ref(false)
const profileError = ref('')
const chatPending = ref(false)
const chatError = ref('')
const sendingMessage = ref(false)
const wsState = ref<WsState>('idle')
const currentUserId = ref('')
const wsClient = ref<WebSocket | null>(null)

const authForm = reactive({
  username: '',
  password: '',
  firstName: '',
})

const profileForm = reactive({
  firstName: profile.value.first_name,
  bio: profile.value.bio ?? '',
  avatar: profile.value.avatar ?? '',
})

const filteredChats = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) {
    return chats.value
  }

  return chats.value.filter((chat) => {
    return (
      chat.title.toLowerCase().includes(query) ||
      chat.username.toLowerCase().includes(query) ||
      chat.id.toLowerCase().includes(query) ||
      chat.subtitle.toLowerCase().includes(query)
    )
  })
})

const selectedChat = computed(() => {
  return chats.value.find((chat) => chat.id === selectedChatId.value) ?? chats.value[0] ?? null
})

const selectedMessages = computed(() => selectedChat.value?.messages ?? [])

const displayName = computed(() => {
  return profile.value.first_name || profile.value.username || 'voidz'
})

const profileBio = computed(() => {
  return profile.value.bio || 'music, code and late nights. building things in the dark.'
})

const connectionLabel = computed(() => {
  if (connection.value === 'connected') {
    return 'backend online'
  }

  if (connection.value === 'offline') {
    return 'backend offline'
  }

  return 'demo mode'
})

const chatListLabel = computed(() => {
  if (!token.value) {
    return 'local demo'
  }

  if (chatPending.value) {
    return 'syncing chats'
  }

  if (wsState.value === 'open') {
    return 'backend online / ws live'
  }

  return `${connectionLabel.value} / ws ${wsState.value}`
})

function makeIcon(icon: IconName) {
  return `<svg viewBox="0 0 24 24" aria-hidden="true">${iconPaths[icon]}</svg>`
}

function shortId(value: string) {
  if (!value) {
    return 'unknown'
  }

  if (value.length <= 12) {
    return value
  }

  return `${value.slice(0, 8)}...${value.slice(-4)}`
}

function conversationTitle(id: string) {
  return `chat ${shortId(id)}`
}

function senderCaption(sender: Sender) {
  if (sender === 'me') {
    return 'you'
  }

  if (sender === 'system') {
    return 'system'
  }

  return shortId(sender)
}

function colorFromId(id: string) {
  let hash = 0

  for (const char of id) {
    hash = (hash * 31 + char.charCodeAt(0)) % 360
  }

  const light = 18 + (hash % 18)
  return `linear-gradient(145deg, hsl(${hash} 7% ${light + 10}%), hsl(${hash} 10% ${light}%) 52%, #070808)`
}

function initials(value: string) {
  return value
    .split(/[\s._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

function formatTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }

  const now = new Date()
  const sameDay = date.toDateString() === now.toDateString()
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)

  if (sameDay) {
    return new Intl.DateTimeFormat('en', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    }).format(date)
  }

  if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  }

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: 'numeric',
  }).format(date)
}

function newestMessage(chat: ChatContact) {
  return chat.messages.at(-1)
}

function updateChatPreview(chat: ChatContact) {
  const latest = newestMessage(chat)
  chat.subtitle = latest?.content ?? (chat.source === 'remote' ? 'no messages yet' : chat.subtitle)
}

function selectChat(chatId: string) {
  selectedChatId.value = chatId
  const chat = chats.value.find((item) => item.id === chatId)
  if (chat) {
    chat.unread = 0
    if (chat.source === 'remote' && !chat.messagesLoaded) {
      void loadMessages(chat.id)
    }
  }
}

function createId(prefix: string) {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return `${prefix}-${crypto.randomUUID()}`
  }

  return `${prefix}-${Date.now()}-${Math.round(Math.random() * 1000)}`
}

function restoreDemoChats() {
  chats.value = demoChatsSnapshot.map((chat) => ({
    ...chat,
    messages: chat.messages.map((message) => ({ ...message })),
  }))
  selectedChatId.value = chats.value[0]?.id ?? ''
}

function createRemoteChat(id: string, previous?: ChatContact): ChatContact {
  return {
    id,
    title: conversationTitle(id),
    username: id,
    subtitle: previous?.subtitle ?? 'open to load messages',
    status: `conversation ${shortId(id)}`,
    accent: colorFromId(id),
    online: previous?.online ?? false,
    unread: previous?.unread ?? 0,
    messages: previous?.messages ?? [],
    source: 'remote',
    messagesLoaded: previous?.messagesLoaded ?? false,
    messagesLoading: previous?.messagesLoading ?? false,
  }
}

function ensureRemoteConversation(id: string) {
  const existing = chats.value.find((chat) => chat.id === id)
  if (existing) {
    if (existing.source !== 'remote') {
      existing.source = 'remote'
    }
    existing.title = conversationTitle(id)
    existing.username = id
    existing.status = `conversation ${shortId(id)}`
    existing.accent = colorFromId(id)
    return existing
  }

  const chat = createRemoteChat(id)
  chats.value = [chat, ...chats.value]
  return chat
}

async function createChat() {
  const targetUsername = newChatUsername.value.trim().replace(/^@+/, '')
  if (!targetUsername) {
    return
  }

  if (token.value && connection.value === 'connected') {
    chatError.value = ''
    chatPending.value = true

    try {
      const response = await firstWorkingApi<string | { conversation_id?: string; conversationId?: string }>(
        conversationPaths(),
        {
          method: 'POST',
          json: {
            target_username: targetUsername,
          },
        },
      )
      const conversationId =
        typeof response === 'string' ? response : response.conversation_id ?? response.conversationId

      if (!conversationId) {
        throw new Error('conversation was not created. Check that target username exists')
      }

      ensureRemoteConversation(conversationId)
      selectedChatId.value = conversationId
      newChatUsername.value = ''
      await loadConversations(true)
      await loadMessages(conversationId, true)
    } catch (error) {
      chatError.value = error instanceof Error ? error.message : 'conversation was not created'
    } finally {
      chatPending.value = false
    }

    return
  }

  const existing = chats.value.find((chat) => chat.username === targetUsername)
  if (existing) {
    selectChat(existing.id)
    newChatUsername.value = ''
    return
  }

  const id = createId(targetUsername)
  const chat: ChatContact = {
    id,
    title: targetUsername,
    username: targetUsername,
    subtitle: 'new local conversation',
    status: 'local',
    accent: 'linear-gradient(145deg, #141414, #3a3a3a 50%, #0b0b0b)',
    online: false,
    unread: 0,
    source: 'demo',
    messagesLoaded: true,
    messages: [
      {
        id: createId('msg'),
        sender: 'system',
        content: `Conversation with ${targetUsername} is ready locally.`,
        createdAt: new Date().toISOString(),
      },
    ],
  }

  chats.value = [chat, ...chats.value]
  selectedChatId.value = id
  newChatUsername.value = ''
}

async function sendMessage() {
  const content = draftMessage.value.trim()
  const chat = selectedChat.value

  if (!content || !chat) {
    return
  }

  if (chat.source === 'remote' && token.value) {
    chatError.value = ''
    sendingMessage.value = true
    draftMessage.value = ''

    try {
      await firstWorkingApi(messagePaths(chat.id), {
        method: 'POST',
        json: {
          conversation_id: chat.id,
          content,
        },
      })
      await loadMessages(chat.id, true)
      await loadConversations(true)
    } catch (error) {
      draftMessage.value = content
      chatError.value = error instanceof Error ? error.message : 'message was not sent'
    } finally {
      sendingMessage.value = false
    }

    return
  }

  const message: ChatMessage = {
    id: createId('msg'),
    sender: 'me',
    content,
    createdAt: new Date().toISOString(),
  }

  chat.messages.push(message)
  updateChatPreview(chat)
  draftMessage.value = ''
}

async function apiRequest<T>(path: string, options: ApiOptions = {}) {
  const headers = new Headers(options.headers)
  const requestOptions: RequestInit = {
    ...options,
    headers,
  }

  if (options.json !== undefined) {
    headers.set('Content-Type', 'application/json')
    requestOptions.body = JSON.stringify(options.json)
  }

  if (token.value) {
    headers.set('Authorization', `Bearer ${token.value}`)
  }

  const response = await fetch(`${API_BASE}${path}`, requestOptions)
  const payload = await response.text()
  let data: unknown = null

  try {
    data = payload ? JSON.parse(payload) : null
  } catch {
    data = payload
  }

  if (!response.ok) {
    const message =
      data && typeof data === 'object' && 'error' in data
        ? String((data as { error: unknown }).error)
        : response.statusText

    throw new ApiError(message || 'Request failed', response.status)
  }

  return data as T
}

function conversationPaths() {
  return [`${CONVERSATIONS_BASE}/`]
}

function messagePaths(conversationId: string) {
  const safeId = encodeURIComponent(conversationId)
  return [`${CONVERSATIONS_BASE}/${safeId}/messages`]
}

async function firstWorkingApi<T>(paths: string[], options: ApiOptions = {}) {
  let lastError: unknown = null

  for (const path of paths) {
    try {
      return await apiRequest<T>(path, options)
    } catch (error) {
      lastError = error

      if (!(error instanceof ApiError) || (error.status !== 404 && error.status !== 405)) {
        throw error
      }
    }
  }

  throw lastError instanceof Error ? lastError : new Error('chat endpoint is unavailable')
}

function normalizeConversationIds(payload: unknown) {
  if (Array.isArray(payload)) {
    return payload.map(String)
  }

  if (payload && typeof payload === 'object') {
    const data = payload as {
      conversations?: unknown
      conversation_ids?: unknown
      data?: unknown
    }

    for (const value of [data.conversations, data.conversation_ids, data.data]) {
      if (Array.isArray(value)) {
        return value.map((item) => {
          if (typeof item === 'string') {
            return item
          }

          if (item && typeof item === 'object') {
            const objectItem = item as { id?: unknown; conversation_id?: unknown; conversationId?: unknown }
            return String(objectItem.id ?? objectItem.conversation_id ?? objectItem.conversationId ?? '')
          }

          return String(item)
        })
      }
    }
  }

  return []
}

function normalizeMessages(payload: unknown, conversationId: string) {
  const rawMessages = Array.isArray(payload)
    ? payload
    : payload && typeof payload === 'object' && Array.isArray((payload as { messages?: unknown }).messages)
      ? (payload as { messages: unknown[] }).messages
      : []

  return rawMessages
    .map((item): ChatMessage | null => {
      if (!item || typeof item !== 'object') {
        return null
      }

      const message = item as Partial<BackendMessage>
      const id = message.id ?? createId('remote-msg')
      const senderId = message.sender_id ?? ''
      const rawCreatedAt = typeof message.created_at === 'string' ? message.created_at : ''
      const createdAt =
        rawCreatedAt && !rawCreatedAt.startsWith('0001-01-01') ? rawCreatedAt : new Date().toISOString()

      return {
        id,
        sender: senderId && senderId === currentUserId.value ? 'me' : senderId || 'system',
        content: message.content ?? '',
        createdAt,
      }
    })
    .filter((message): message is ChatMessage => Boolean(message && message.content))
    .map((message) => ({
      ...message,
      id: `${conversationId}-${message.id}`,
    }))
    .sort((left, right) => new Date(left.createdAt).getTime() - new Date(right.createdAt).getTime())
}

async function loadConversations(quiet = false) {
  if (!token.value) {
    return
  }

  if (!quiet) {
    chatPending.value = true
  }

  try {
    const payload = await firstWorkingApi<unknown>(conversationPaths())
    const ids = normalizeConversationIds(payload).filter(Boolean)
    const previous = new Map(chats.value.map((chat) => [chat.id, chat]))
    const remoteChats = ids.map((id) => createRemoteChat(id, previous.get(id)))

    chats.value = remoteChats
    chatError.value = ''

    if (!remoteChats.some((chat) => chat.id === selectedChatId.value)) {
      selectedChatId.value = remoteChats[0]?.id ?? ''
    }

    if (selectedChatId.value) {
      await loadMessages(selectedChatId.value, true)
    }
  } catch (error) {
    if (!quiet) {
      chatError.value = error instanceof Error ? error.message : 'chat list was not loaded'
    }
  } finally {
    chatPending.value = false
  }
}

async function loadMessages(conversationId: string, quiet = false) {
  if (!token.value || !conversationId) {
    return
  }

  const chat = ensureRemoteConversation(conversationId)
  chat.messagesLoading = true

  try {
    const payload = await firstWorkingApi<unknown>(messagePaths(conversationId))
    chat.messages = normalizeMessages(payload, conversationId)
    chat.messagesLoaded = true
    chat.unread = selectedChatId.value === conversationId ? 0 : chat.unread
    updateChatPreview(chat)

    if (!quiet) {
      chatError.value = ''
    }
  } catch (error) {
    if (!quiet) {
      chatError.value = error instanceof Error ? error.message : 'messages were not loaded'
    }
  } finally {
    chat.messagesLoading = false
  }
}

function decodeJwtSubject(jwtToken: string) {
  try {
    const payload = jwtToken.split('.')[1]
    if (!payload) {
      return ''
    }

    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded = JSON.parse(atob(normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '='))) as {
      sub?: unknown
    }

    return typeof decoded.sub === 'string' ? decoded.sub : ''
  } catch {
    return ''
  }
}

function buildWsUrl() {
  const url = new URL(API_BASE || window.location.origin, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = WS_PATH
  url.searchParams.set('token', token.value)
  return url.toString()
}

async function handleSocketEvent(event: WsEvent) {
  if (event.event_type !== 'new_message') {
    return
  }

  const conversationId = event.payload?.conversation_id
  if (!conversationId) {
    await loadConversations(true)
    return
  }

  const chat = ensureRemoteConversation(conversationId)
  if (selectedChatId.value !== conversationId) {
    chat.unread += 1
  }

  await loadConversations(true)
  await loadMessages(conversationId, true)
}

function connectWebSocket() {
  disconnectWebSocket()

  if (!token.value || typeof WebSocket === 'undefined') {
    wsState.value = 'idle'
    return
  }

  wsState.value = 'connecting'

  try {
    const socket = new WebSocket(buildWsUrl())
    wsClient.value = socket

    socket.addEventListener('open', () => {
      wsState.value = 'open'
    })

    socket.addEventListener('message', (messageEvent) => {
      try {
        const event = JSON.parse(String(messageEvent.data)) as WsEvent
        void handleSocketEvent(event)
      } catch {
        wsState.value = 'error'
      }
    })

    socket.addEventListener('error', () => {
      wsState.value = 'error'
    })

    socket.addEventListener('close', () => {
      if (wsClient.value === socket) {
        wsClient.value = null
        wsState.value = wsState.value === 'error' ? 'error' : 'closed'
      }
    })
  } catch {
    wsState.value = 'error'
  }
}

function disconnectWebSocket() {
  const socket = wsClient.value
  if (!socket) {
    return
  }

  wsClient.value = null
  socket.close()
}

async function loadMe() {
  if (!token.value) {
    connection.value = 'demo'
    return
  }

  currentUserId.value = decodeJwtSubject(token.value)

  try {
    const response = await apiRequest<GetMeResponse>('/api/v1/users/me')
    profile.value = response.user
    profileForm.firstName = response.user.first_name
    profileForm.bio = response.user.bio ?? ''
    profileForm.avatar = response.user.avatar ?? ''
    connection.value = 'connected'
  } catch {
    connection.value = 'offline'
  }
}

async function submitAuth() {
  authError.value = ''

  if (!authForm.username.trim() || !authForm.password.trim()) {
    authError.value = 'username and password are required'
    return
  }

  authPending.value = true

  try {
    if (authMode.value === 'register') {
      await apiRequest<{ id: string }>('/api/v1/auth/register', {
        method: 'POST',
        json: {
          username: authForm.username.trim(),
          password: authForm.password,
          first_name: authForm.firstName.trim() || authForm.username.trim(),
        },
      })
    }

    const response = await apiRequest<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      json: {
        username: authForm.username.trim(),
        password: authForm.password,
      },
    })

    token.value = response.token
    localStorage.setItem(TOKEN_STORAGE_KEY, response.token)
    authOpen.value = false
    await loadMe()
    await loadConversations()
    connectWebSocket()
  } catch (error) {
    connection.value = 'offline'
    authError.value = error instanceof Error ? error.message : 'backend is unavailable'
  } finally {
    authPending.value = false
  }
}

async function saveProfile() {
  profileError.value = ''

  if (!token.value) {
    profile.value = {
      ...profile.value,
      first_name: profileForm.firstName.trim() || profile.value.username,
      bio: profileForm.bio.trim() || null,
      avatar: profileForm.avatar.trim() || null,
    }
    profileOpen.value = false
    return
  }

  profilePending.value = true

  try {
    await apiRequest('/api/v1/users/me', {
      method: 'PUT',
      json: {
        first_name: profileForm.firstName.trim() || profile.value.username,
        bio: profileForm.bio.trim() || null,
        avatar: profileForm.avatar.trim() || null,
      },
    })
    await loadMe()
    profileOpen.value = false
  } catch (error) {
    profileError.value = error instanceof Error ? error.message : 'profile was not saved'
  } finally {
    profilePending.value = false
  }
}

function logout() {
  disconnectWebSocket()
  token.value = ''
  currentUserId.value = ''
  localStorage.removeItem(TOKEN_STORAGE_KEY)
  connection.value = 'demo'
  wsState.value = 'idle'
  chatError.value = ''
  restoreDemoChats()
}

function openAuth(mode: AuthMode) {
  authMode.value = mode
  authError.value = ''
  authOpen.value = true
}

onMounted(async () => {
  currentUserId.value = decodeJwtSubject(token.value)
  await loadMe()

  if (token.value && connection.value === 'connected') {
    await loadConversations()
    connectWebSocket()
  }
})

onBeforeUnmount(disconnectWebSocket)
</script>

<template>
  <main class="desktop">
    <section class="window" aria-label="Pooky messenger">
      <aside class="rail" aria-label="Navigation">
        <button class="brand-button" type="button" aria-label="Pooky">
          <span v-html="makeIcon('message')"></span>
        </button>

        <nav class="rail-nav">
          <button
            v-for="item in navItems"
            :key="item.id"
            class="rail-button"
            :class="{ active: item.id === 'inbox' }"
            type="button"
            :aria-label="item.label"
            :title="item.label"
          >
            <span v-html="makeIcon(item.icon)"></span>
          </button>
        </nav>

        <button class="rail-button rail-bottom" type="button" aria-label="Log out" title="Log out" @click="logout">
          <span v-html="makeIcon('logout')"></span>
        </button>
      </aside>

      <section class="profile-panel" aria-label="Profile">
        <div class="profile-card">
          <div class="avatar avatar-large" :class="{ online: connection !== 'offline' }">
            <img v-if="profile.avatar" :src="profile.avatar" :alt="displayName" />
            <span v-else>{{ initials(displayName) }}</span>
          </div>

          <h1>{{ displayName }}</h1>
          <p class="handle">@{{ profile.username }}</p>
          <p class="bio">{{ profileBio }}</p>

          <div class="profile-actions">
            <button class="icon-button" type="button" aria-label="Edit profile" title="Edit profile" @click="profileOpen = true">
              <span v-html="makeIcon('user')"></span>
            </button>
            <button class="quiet-button" type="button" @click="openAuth('login')">
              {{ token ? connectionLabel : 'Connect backend' }}
            </button>
          </div>
        </div>

        <div class="player-panel">
          <div class="panel-heading">
            <div class="service-title">
              <span class="service-icon" v-html="makeIcon('music')"></span>
              <span>Spotify</span>
            </div>
            <button class="link-button" type="button" aria-label="Open in Spotify" title="Open in Spotify">
              <span>Open in Spotify</span>
              <span v-html="makeIcon('open')"></span>
            </button>
          </div>

          <article class="track-card">
            <div class="album-art">
              <span></span>
            </div>
            <div class="track-copy">
              <strong>The Night We Met</strong>
              <span>Lord Huron</span>
            </div>

            <div class="progress">
              <div class="progress-bar">
                <span></span>
              </div>
              <div class="track-time">
                <span>1:42</span>
                <span>3:28</span>
              </div>
            </div>

            <div class="player-controls">
              <button class="icon-button bare" type="button" aria-label="Like" title="Like">
                <span v-html="makeIcon('heart')"></span>
              </button>
              <button class="icon-button bare" type="button" aria-label="Previous" title="Previous">
                <span v-html="makeIcon('previous')"></span>
              </button>
              <button class="play-button" type="button" aria-label="Pause" title="Pause">
                <span v-html="makeIcon('pause')"></span>
              </button>
              <button class="icon-button bare" type="button" aria-label="Next" title="Next">
                <span v-html="makeIcon('next')"></span>
              </button>
              <button class="spotify-dot" type="button" aria-label="Spotify" title="Spotify">
                <span></span>
              </button>
            </div>
          </article>
        </div>
      </section>

      <section class="workspace" aria-label="Messenger workspace">
        <header class="window-bar">
          <span></span>
          <div class="window-controls" aria-hidden="true">
            <i></i>
            <i></i>
            <i></i>
          </div>
        </header>

        <div class="chat-board">
          <section class="chat-list-panel" aria-label="Chats">
            <div class="chat-list-header">
              <div>
                <p class="eyebrow">{{ chatListLabel }}</p>
                <h2>Chats</h2>
              </div>

              <form class="new-chat-form" @submit.prevent="createChat">
                <label class="sr-only" for="new-chat">Target username</label>
                <input id="new-chat" v-model="newChatUsername" type="text" placeholder="target username" autocomplete="off" />
                <button class="icon-button" type="submit" aria-label="Create chat" title="Create chat">
                  <span v-html="makeIcon('plus')"></span>
                </button>
              </form>
            </div>

            <label class="search-box" for="chat-search">
              <span v-html="makeIcon('search')"></span>
              <input id="chat-search" v-model="searchQuery" type="search" placeholder="Search" autocomplete="off" />
            </label>

            <p v-if="chatError" class="sync-note error">{{ chatError }}</p>

            <div class="chat-list">
              <button
                v-for="chat in filteredChats"
                :key="chat.id"
                class="chat-row"
                :class="{ active: selectedChatId === chat.id, remote: chat.source === 'remote' }"
                type="button"
                @click="selectChat(chat.id)"
              >
                <span class="avatar avatar-small" :class="{ online: chat.online }" :style="{ background: chat.accent }">
                  <img v-if="chat.avatar" :src="chat.avatar" :alt="chat.title" />
                  <span v-else>{{ chat.source === 'remote' ? '#' : initials(chat.title) }}</span>
                </span>

                <span class="chat-row-main">
                  <strong>{{ chat.title }}</strong>
                  <small>{{ chat.subtitle }}</small>
                </span>

                <span class="chat-meta">
                  <time>{{ formatTime(newestMessage(chat)?.createdAt ?? '') }}</time>
                  <span v-if="chat.unread" class="unread">{{ chat.unread }}</span>
                </span>
              </button>

              <div v-if="!filteredChats.length" class="empty-state">
                <strong>No conversations</strong>
                <span>{{ token ? 'Create one or wait for backend conversations.' : 'Connect backend or create a local chat.' }}</span>
              </div>
            </div>
          </section>

          <section v-if="selectedChat" class="conversation-panel" aria-label="Conversation">
            <header class="conversation-header">
              <div class="conversation-person">
                <span
                  class="avatar avatar-medium"
                  :class="{ online: selectedChat.online }"
                  :style="{ background: selectedChat.accent }"
                >
                  <img v-if="selectedChat.avatar" :src="selectedChat.avatar" :alt="selectedChat.title" />
                  <span v-else>{{ selectedChat.source === 'remote' ? '#' : initials(selectedChat.title) }}</span>
                </span>
                <div>
                  <h2>{{ selectedChat.title }}</h2>
                  <p>{{ selectedChat.status }}</p>
                </div>
              </div>

              <div class="conversation-actions">
                <button class="icon-button" type="button" aria-label="Call" title="Call">
                  <span v-html="makeIcon('phone')"></span>
                </button>
                <button class="icon-button" type="button" aria-label="Photo" title="Photo">
                  <span v-html="makeIcon('image')"></span>
                </button>
                <button class="icon-button" type="button" aria-label="More" title="More">
                  <span v-html="makeIcon('more')"></span>
                </button>
              </div>
            </header>

            <div class="messages">
              <div v-if="selectedChat.messagesLoading" class="empty-state compact">
                <strong>Loading messages</strong>
              </div>

              <article
                v-for="message in selectedMessages"
                :key="message.id"
                class="message"
                :class="{ mine: message.sender === 'me', system: message.sender === 'system' }"
              >
                <span v-if="message.sender !== 'me' && message.sender !== 'system'" class="message-sender">
                  {{ senderCaption(message.sender) }}
                </span>
                <p>{{ message.content }}</p>
                <time>{{ formatTime(message.createdAt) }}</time>
              </article>

              <div v-if="!selectedChat.messagesLoading && !selectedMessages.length" class="empty-state compact">
                <strong>No messages yet</strong>
              </div>
            </div>

            <form class="composer" @submit.prevent="sendMessage">
              <label class="sr-only" for="message-draft">Message</label>
              <input
                id="message-draft"
                v-model="draftMessage"
                type="text"
                placeholder="Message"
                autocomplete="off"
                :disabled="sendingMessage"
                @keydown.enter.exact.prevent="sendMessage"
              />
              <button class="send-button" type="submit" aria-label="Send" title="Send" :disabled="sendingMessage">
                <span v-html="makeIcon('send')"></span>
              </button>
            </form>
          </section>

          <section v-else class="conversation-panel empty-conversation" aria-label="Conversation">
            <div class="empty-state">
              <strong>Select a chat</strong>
              <span>{{ token ? 'Backend conversations will appear by id.' : 'Demo mode is active until you sign in.' }}</span>
            </div>
          </section>
        </div>
      </section>
    </section>

    <div v-if="authOpen" class="modal-layer" role="dialog" aria-modal="true" aria-labelledby="auth-title">
      <form class="modal-card" @submit.prevent="submitAuth">
        <button class="icon-button modal-close" type="button" aria-label="Close" title="Close" @click="authOpen = false">
          <span v-html="makeIcon('x')"></span>
        </button>
        <p class="eyebrow">{{ authMode }}</p>
        <h2 id="auth-title">{{ authMode === 'login' ? 'Welcome back' : 'Create account' }}</h2>

        <label>
          <span>Username</span>
          <input v-model="authForm.username" type="text" autocomplete="username" />
        </label>

        <label v-if="authMode === 'register'">
          <span>First name</span>
          <input v-model="authForm.firstName" type="text" autocomplete="given-name" />
        </label>

        <label>
          <span>Password</span>
          <input v-model="authForm.password" type="password" autocomplete="current-password" />
        </label>

        <p v-if="authError" class="form-error">{{ authError }}</p>

        <button class="primary-button" type="submit" :disabled="authPending">
          {{ authPending ? 'Please wait' : authMode === 'login' ? 'Sign in' : 'Register' }}
        </button>

        <button
          class="text-button"
          type="button"
          @click="authMode = authMode === 'login' ? 'register' : 'login'"
        >
          {{ authMode === 'login' ? 'Need an account?' : 'I have an account' }}
        </button>
      </form>
    </div>

    <div v-if="profileOpen" class="modal-layer" role="dialog" aria-modal="true" aria-labelledby="profile-title">
      <form class="modal-card" @submit.prevent="saveProfile">
        <button class="icon-button modal-close" type="button" aria-label="Close" title="Close" @click="profileOpen = false">
          <span v-html="makeIcon('x')"></span>
        </button>
        <p class="eyebrow">{{ connectionLabel }}</p>
        <h2 id="profile-title">Profile</h2>

        <label>
          <span>First name</span>
          <input v-model="profileForm.firstName" type="text" autocomplete="given-name" />
        </label>

        <label>
          <span>Bio</span>
          <textarea v-model="profileForm.bio" rows="3"></textarea>
        </label>

        <label>
          <span>Avatar URL</span>
          <input v-model="profileForm.avatar" type="url" autocomplete="off" />
        </label>

        <p v-if="profileError" class="form-error">{{ profileError }}</p>

        <button class="primary-button" type="submit" :disabled="profilePending">
          {{ profilePending ? 'Saving' : 'Save' }}
        </button>
      </form>
    </div>
  </main>
</template>
