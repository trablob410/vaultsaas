'use client'

import type { PolicyTemplateVersion } from '@/types/api'
import { Card, CardContent } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Props {
  open: boolean
  name: string
  versions: PolicyTemplateVersion[]
  onOpenChange: (open: boolean) => void
}

export function PolicyTemplateHistoryDialog({ open, name, versions, onOpenChange }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Version history — {name}</DialogTitle>
        </DialogHeader>
        {versions.length === 0 ? (
          <p className="text-sm text-muted-foreground">No versions found.</p>
        ) : (
          <div className="space-y-2 max-h-[50vh] overflow-y-auto">
            {versions.map((v) => (
              <Card key={v.id}>
                <CardContent className="pt-4 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">Version {v.version}</span>
                    <span className="text-muted-foreground">{new Date(v.created_at).toLocaleString()}</span>
                  </div>
                  {v.change_note && <p className="text-muted-foreground mt-1">{v.change_note}</p>}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
