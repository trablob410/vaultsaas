export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-8">
      <div className="text-center">
        <h1 className="text-4xl font-bold tracking-tight">Valt</h1>
        <p className="mt-4 text-lg text-gray-600">
          AI Secret Vault with Human-in-the-Loop Approval
        </p>
        <div className="mt-8 flex gap-4 justify-center">
          <a
            href="/login"
            className="rounded-lg bg-black px-6 py-3 text-sm font-medium text-white hover:bg-gray-800"
          >
            Sign In
          </a>
          <a
            href="/register"
            className="rounded-lg border border-gray-300 px-6 py-3 text-sm font-medium hover:bg-gray-50"
          >
            Get Started
          </a>
        </div>
      </div>
    </main>
  );
}
