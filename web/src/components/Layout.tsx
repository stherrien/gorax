import { Link, Outlet, useLocation } from 'react-router-dom'

const navigation = [
  { name: 'Dashboard', href: '/' },
  { name: 'Workflows', href: '/workflows' },
  { name: 'Executions', href: '/executions' },
]

export default function Layout() {
  const location = useLocation()

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 w-64 bg-gray-800 border-r border-gray-700">
        {/* Logo */}
        <div className="flex items-center h-16 px-6 border-b border-gray-700">
          <span className="text-2xl font-bold text-white">gorax</span>
        </div>

        {/* Navigation */}
        <nav className="mt-6 px-3">
          {navigation.map((item) => {
            const isActive = location.pathname === item.href ||
              (item.href !== '/' && location.pathname.startsWith(item.href))

            return (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  flex items-center px-4 py-3 mb-1 rounded-lg text-sm font-medium transition-colors
                  ${isActive
                    ? 'bg-primary-600 text-white'
                    : 'text-gray-300 hover:bg-gray-700 hover:text-white'}
                `}
              >
                {item.name}
              </Link>
            )
          })}
        </nav>
      </div>

      {/* Main content */}
      <div className="pl-64">
        {/* Header */}
        <header className="h-16 bg-gray-800 border-b border-gray-700 flex items-center justify-between px-6">
          <div></div>
          <div className="flex items-center space-x-4">
            <Link
              to="/workflows/new"
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
            >
              New Workflow
            </Link>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
