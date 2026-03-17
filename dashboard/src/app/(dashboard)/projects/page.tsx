'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api-client'
import type { Project } from '@/types/api'
import { FolderOpen } from 'lucide-react'
import Link from 'next/link'

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const workspaceId = typeof window !== 'undefined'
      ? localStorage.getItem('valt_current_workspace')
      : null

    if (!workspaceId) {
      setLoading(false)
      return
    }

    api.projects.list(workspaceId)
      .then(r => setProjects(r.projects ?? []))
      .catch(e => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [])

  const hasWorkspace = typeof window !== 'undefined' && !!localStorage.getItem('valt_current_workspace')

  if (!hasWorkspace) {
    return (
      <div className="p-6 max-w-xl">
        <div className="flex items-center gap-3 mb-6">
          <FolderOpen className="w-5 h-5 text-muted-foreground" />
          <h1 className="text-xl font-semibold">Projects</h1>
        </div>
        <div className="p-6 border rounded-lg bg-card text-center space-y-3">
          <FolderOpen className="w-8 h-8 text-muted-foreground mx-auto" />
          <p className="text-sm font-medium">No workspace selected</p>
          <p className="text-xs text-muted-foreground">
            Navigate to an organization and workspace first to view its projects.
          </p>
          <Link
            href="/orgs"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
          >
            Browse Organizations
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center gap-3 mb-6">
        <FolderOpen className="w-5 h-5 text-muted-foreground" />
        <h1 className="text-xl font-semibold">Projects</h1>
      </div>

      {error && <p className="mb-4 text-sm text-destructive">{error}</p>}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : projects.length === 0 ? (
        <p className="text-sm text-muted-foreground">No projects in this workspace yet.</p>
      ) : (
        <ul className="space-y-2">
          {projects.map(project => (
            <li key={project.id} className="p-4 border rounded-lg bg-card hover:bg-accent/50 transition-colors">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-md bg-primary/10 flex items-center justify-center shrink-0">
                  <FolderOpen className="w-4 h-4 text-primary" />
                </div>
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{project.name}</p>
                  <p className="text-xs text-muted-foreground font-mono">{project.slug}</p>
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
