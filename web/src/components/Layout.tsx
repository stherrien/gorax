import { Link, Outlet, useLocation } from 'react-router-dom'
import { useThemeContext } from '../contexts/ThemeContext'

const navigation = [
  { name: 'Dashboard', href: '/', adminOnly: false },
  { name: 'Workflows', href: '/workflows', adminOnly: false },
  { name: 'Marketplace', href: '/marketplace', adminOnly: false },
  { name: 'AI Builder', href: '/ai/builder', adminOnly: false },
  { name: 'Webhooks', href: '/webhooks', adminOnly: false },
  { name: 'Schedules', href: '/schedules', adminOnly: false },
  { name: 'Executions', href: '/executions', adminOnly: false },
  { name: 'Analytics', href: '/analytics', adminOnly: false },
  { name: 'Credentials', href: '/credentials', adminOnly: false },
  { name: 'OAuth', href: '/oauth/connections', adminOnly: false },
  { name: 'Docs', href: '/docs', adminOnly: false },
]

const adminNavigation = [
  { name: 'SSO Settings', href: '/admin/sso' },
  { name: 'Audit Logs', href: '/admin/audit-logs' },
]

export default function Layout() {
  const location = useLocation()
  const { theme, isDark, toggleTheme } = useThemeContext()

  // TODO: Replace with actual user role from auth context
  // For now, check localStorage for dev mode
  const isAdmin = localStorage.getItem('user_role') === 'admin' || true // Enable for dev

  return (
    <div className={`min-h-screen ${isDark ? 'bg-gray-900' : 'bg-gray-100'}`}>
      {/* Sidebar */}
      <div className={`fixed inset-y-0 left-0 w-64 ${isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border-r`}>
        {/* Logo */}
        <div className={`flex items-center h-16 px-6 border-b ${isDark ? 'border-gray-700' : 'border-gray-200'}`}>
          <span className={`text-2xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>gorax</span>
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
                    : isDark
                      ? 'text-gray-300 hover:bg-gray-700 hover:text-white'
                      : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}
                `}
              >
                {item.name}
              </Link>
            )
          })}

          {/* Admin Section */}
          {isAdmin && (
            <>
              <div className={`mt-6 mb-2 px-4 text-xs font-semibold uppercase tracking-wider ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
                Admin
              </div>
              {adminNavigation.map((item) => {
                const isActive = location.pathname === item.href ||
                  location.pathname.startsWith(item.href)

                return (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`
                      flex items-center px-4 py-3 mb-1 rounded-lg text-sm font-medium transition-colors
                      ${isActive
                        ? 'bg-primary-600 text-white'
                        : isDark
                          ? 'text-gray-300 hover:bg-gray-700 hover:text-white'
                          : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}
                    `}
                  >
                    {item.name}
                  </Link>
                )
              })}
            </>
          )}
        </nav>
      </div>

      {/* Main content */}
      <div className="pl-64">
        {/* Header */}
        <header className={`h-16 ${isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border-b flex items-center justify-between px-6`}>
          <div></div>
          <div className="flex items-center space-x-4">
            {/* Theme Toggle */}
            <button
              onClick={toggleTheme}
              className={`p-2 rounded-lg transition-colors ${isDark ? 'text-gray-400 hover:text-white hover:bg-gray-700' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100'}`}
              aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
                  />
                </svg>
              ) : (
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
                  />
                </svg>
              )}
            </button>
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
