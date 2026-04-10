import { useEffect } from "react";
import { Routes, Route } from "react-router-dom";
import { useAuthStore } from "@/lib/auth";
import Layout from "@/components/Layout";
import ProtectedRoute from "@/components/ProtectedRoute";
import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import ItineraryList from "@/pages/itineraries/ItineraryList";
import ItineraryWizard from "@/pages/itineraries/ItineraryWizard";
import ItineraryDetail from "@/pages/itineraries/ItineraryDetail";
import BookingList from "@/pages/BookingList";
import BookingNew from "@/pages/BookingNew";
import BookingDetail from "@/pages/BookingDetail";
import ProcurementDashboard from "@/pages/ProcurementDashboard";
import NotificationCenter from "@/pages/NotificationCenter";
import WalletDashboard from "@/pages/WalletDashboard";
import DocumentCenter from "@/pages/DocumentCenter";
import ReviewDashboard from "@/pages/ReviewDashboard";
import AdminLayout from "@/pages/admin/AdminLayout";
import AdminOverview from "@/pages/admin/AdminOverview";
import AdminUsers from "@/pages/admin/AdminUsers";
import AdminSettings from "@/pages/admin/AdminSettings";

export default function App() {
  const { hydrate } = useAuthStore();

  useEffect(() => {
    hydrate();
  }, [hydrate]);

  return (
    <Routes>
      <Route path="/login" element={<Login />} />

      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="itineraries" element={<ItineraryList />} />
        <Route path="itineraries/new" element={<ItineraryWizard />} />
        <Route path="itineraries/:id" element={<ItineraryDetail />} />
        <Route path="bookings" element={<BookingList />} />
        <Route path="bookings/new" element={<BookingNew />} />
        <Route path="bookings/:id" element={<BookingDetail />} />
        <Route path="procurement" element={<ProcurementDashboard />} />
        <Route path="notifications" element={<NotificationCenter />} />
        <Route path="wallet" element={<WalletDashboard />} />
        <Route path="documents" element={<DocumentCenter />} />
        <Route path="reviews" element={<ReviewDashboard />} />

        <Route
          path="admin"
          element={
            <ProtectedRoute requiredRoles={["administrator"]}>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<AdminOverview />} />
          <Route path="users" element={<AdminUsers />} />
          <Route path="settings" element={<AdminSettings />} />
        </Route>
      </Route>
    </Routes>
  );
}
