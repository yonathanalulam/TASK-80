import { NavLink, Outlet } from "react-router-dom";

const adminNavItems = [
  { to: "/admin", label: "Overview", end: true },
  { to: "/admin/users", label: "Users", end: false },
  { to: "/admin/settings", label: "Settings", end: false },
];

export default function AdminLayout() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Administration</h1>
        <p className="mt-1 text-gray-500">Manage users, settings, and more.</p>
      </div>

      <nav className="flex gap-4 border-b border-gray-200">
        {adminNavItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            className={({ isActive }) =>
              `border-b-2 pb-3 text-sm font-medium transition-colors ${
                isActive
                  ? "border-indigo-600 text-indigo-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`
            }
          >
            {item.label}
          </NavLink>
        ))}
      </nav>

      <Outlet />
    </div>
  );
}
