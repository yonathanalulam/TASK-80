import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuthStore } from "@/lib/auth";

const navItems = [
  { to: "/", label: "Dashboard", icon: "grid" },
  { to: "/itineraries", label: "Itineraries", icon: "map" },
  { to: "/bookings", label: "Bookings", icon: "calendar" },
  { to: "/procurement", label: "Procurement", icon: "truck" },
  { to: "/notifications", label: "Messages", icon: "bell" },
  { to: "/wallet", label: "Wallet", icon: "wallet" },
  { to: "/documents", label: "Documents", icon: "file" },
  { to: "/reviews", label: "Reviews", icon: "star" },
];

const iconMap: Record<string, string> = {
  grid: "\u25A6",
  map: "\u25B2",
  calendar: "\u25A3",
  truck: "\u25B6",
  bell: "\u25CF",
  wallet: "\u25A0",
  file: "\u25AD",
  star: "\u2605",
  shield: "\u25C6",
};

export default function Layout() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const isAdmin = user?.roles?.includes("administrator") ?? false;

  return (
    <div className="flex h-screen bg-gray-50">
      <aside className="hidden w-64 flex-shrink-0 flex-col border-r border-gray-200 bg-white md:flex">
        <div className="flex h-16 items-center border-b border-gray-200 px-6">
          <h1 className="text-xl font-bold text-indigo-600">TrailForge</h1>
        </div>

        <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-indigo-50 text-indigo-700"
                    : "text-gray-700 hover:bg-gray-100"
                }`
              }
            >
              <span className="text-base">{iconMap[item.icon]}</span>
              {item.label}
            </NavLink>
          ))}

          {isAdmin && (
            <NavLink
              to="/admin"
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-indigo-50 text-indigo-700"
                    : "text-gray-700 hover:bg-gray-100"
                }`
              }
            >
              <span className="text-base">{iconMap.shield}</span>
              Admin
            </NavLink>
          )}
        </nav>

        <div className="border-t border-gray-200 p-4">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-100 text-sm font-medium text-indigo-700">
              {user?.name?.charAt(0).toUpperCase() ?? "?"}
            </div>
            <div className="flex-1 truncate">
              <p className="truncate text-sm font-medium text-gray-900">
                {user?.name ?? "User"}
              </p>
              <p className="truncate text-xs text-gray-500">
                {user?.email ?? ""}
              </p>
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="mt-3 w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-gray-100"
          >
            Sign out
          </button>
        </div>
      </aside>

      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
          <h2 className="text-lg font-semibold text-gray-900">TrailForge</h2>
          <div className="flex items-center gap-4">
            <NavLink
              to="/notifications"
              className="text-gray-500 hover:text-gray-700"
            >
              {iconMap.bell}
            </NavLink>
            <span className="text-sm text-gray-600">{user?.name}</span>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
