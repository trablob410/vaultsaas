'use client'

import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type { PolicyParameters } from '@/types/api'
import { PolicyParameterFields } from './policy-parameter-fields'

type Mode = 'create' | 'edit' | 'clone'

interface Props {
  open: boolean
  mode: Mode
  name: string
  description: string
  changeNote: string
  params: PolicyParameters
  onOpenChange: (open: boolean) => void
  onNameChange: (value: string) => void
  onDescriptionChange: (value: string) => void
  onChangeNoteChange: (value: string) => void
  onParamsChange: (value: PolicyParameters) => void
  onSubmit: (e: React.FormEvent) => void
}

export function PolicyTemplateEditorDialog(props: Props) {
  const {
    open,
    mode,
    name,
    description,
    changeNote,
    params,
    onOpenChange,
    onNameChange,
    onDescriptionChange,
    onChangeNoteChange,
    onParamsChange,
    onSubmit,
  } = props

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{mode === 'create' ? 'Create template' : mode === 'clone' ? 'Clone template' : 'Edit template (new version)'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          {(mode === 'create' || mode === 'clone') && (
            <div className="space-y-1.5">
              <Label htmlFor="tpl-name">Name</Label>
              <Input id="tpl-name" value={name} onChange={(e) => onNameChange(e.target.value)} required />
            </div>
          )}
          {mode === 'create' && (
            <div className="space-y-1.5">
              <Label htmlFor="tpl-desc">Description</Label>
              <Textarea id="tpl-desc" rows={2} value={description} onChange={(e) => onDescriptionChange(e.target.value)} />
            </div>
          )}
          {mode === 'clone' && (
            <p className="text-sm text-muted-foreground">
              Clone creates a new custom template from current source version.
            </p>
          )}
          {mode === 'edit' && (
            <div className="space-y-1.5">
              <Label htmlFor="change-note">Change note</Label>
              <Input id="change-note" value={changeNote} onChange={(e) => onChangeNoteChange(e.target.value)} placeholder="What changed in this version?" />
            </div>
          )}
          {mode !== 'clone' && <PolicyParameterFields value={params} onChange={onParamsChange} />}
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit">Save</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
