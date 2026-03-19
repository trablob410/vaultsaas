'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Trash2 } from 'lucide-react'
import { api } from '@/lib/api-client'
import type { NotificationChannel } from '@/types/api'

const CHANNEL_LABELS: Record<string, string> = {
  email: 'Email',
  slack: 'Slack User ID',
  telegram: 'Telegram Chat ID',
}

export default function NotificationsPage() {
  const [channels, setChannels] = useState<NotificationChannel[]>([])
  const [channelType, setChannelType] = useState('email')
  const [handle, setHandle] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.notificationChannels.list().then(res => setChannels(res.channels)).catch(() => {})
  }, [])

  async function handleAdd() {
    if (!handle.trim()) return
    setLoading(true)
    setError('')
    try {
      const ch = await api.notificationChannels.upsert(channelType, handle.trim())
      setChannels(prev => {
        const filtered = prev.filter(c => c.channel_type !== ch.channel_type)
        return [...filtered, ch]
      })
      setHandle('')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    try {
      await api.notificationChannels.delete(id)
      setChannels(prev => prev.filter(c => c.id !== id))
    } catch {
      setError('Failed to remove channel')
    }
  }

  return (
    <div className="max-w-lg space-y-4">
      <h2 className="text-lg font-semibold">Notification Channels</h2>
      <p className="text-sm text-muted-foreground">
        Link channels to receive approval notifications. Slack and Telegram require verification after linking.
      </p>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Add Channel</CardTitle>
          <CardDescription>One channel per type. Adding again replaces the existing one.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex gap-2">
            <Select value={channelType} onValueChange={setChannelType}>
              <SelectTrigger className="w-36">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="email">Email</SelectItem>
                <SelectItem value="slack">Slack</SelectItem>
                <SelectItem value="telegram">Telegram</SelectItem>
              </SelectContent>
            </Select>
            <Input
              placeholder={CHANNEL_LABELS[channelType]}
              value={handle}
              onChange={e => setHandle(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleAdd()}
              className="flex-1"
            />
            <Button onClick={handleAdd} disabled={loading || !handle.trim()}>
              {loading ? 'Saving…' : 'Save'}
            </Button>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </CardContent>
      </Card>

      {channels.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-base">Linked Channels</CardTitle></CardHeader>
          <CardContent className="space-y-2">
            {channels.map(ch => (
              <div key={ch.id} className="flex items-center justify-between rounded-md border px-3 py-2">
                <div className="space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium capitalize">{ch.channel_type}</span>
                    {ch.verified
                      ? <Badge variant="default" className="text-xs">Verified</Badge>
                      : <Badge variant="secondary" className="text-xs">Pending verification</Badge>
                    }
                  </div>
                  <p className="text-xs text-muted-foreground font-mono">{ch.handle}</p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-muted-foreground hover:text-destructive"
                  onClick={() => handleDelete(ch.id)}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
