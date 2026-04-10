import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "@/lib/api";

interface FileRecord {
  id: string;
  originalFilename: string;
  mimeType: string;
  byteSize: number;
  encrypted: boolean;
  ownerUserId: string;
  createdAt: string;
}

interface ContractTemplate {
  id: string;
  name: string;
  description: string;
  version: string;
}

interface InvoiceRequest {
  id: string;
  requesterId: string;
  orderType: string;
  orderId: string;
  status: string;
  notes: string;
  createdAt: string;
}

const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  generated: "bg-green-100 text-green-700",
  sent: "bg-blue-100 text-blue-700",
  paid: "bg-green-100 text-green-700",
  cancelled: "bg-red-100 text-red-700",
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function statusLabel(status: string): string {
  return status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export default function DocumentCenter() {
  const [dragOver, setDragOver] = useState(false);
  const [showInvoiceModal, setShowInvoiceModal] = useState(false);
  const [invoiceOrderType, setInvoiceOrderType] = useState("booking");
  const [invoiceOrderId, setInvoiceOrderId] = useState("");
  const [invoiceNotes, setInvoiceNotes] = useState("");
  const qc = useQueryClient();

  const [uploadRecordType, setUploadRecordType] = useState("booking");
  const [uploadRecordId, setUploadRecordId] = useState("");

  const { data: files, isLoading: filesLoading } = useQuery<FileRecord[]>({
    queryKey: ["files", uploadRecordType, uploadRecordId],
    queryFn: async () => {
      const { data } = await api.get(`/files/record/${uploadRecordType}/${uploadRecordId}`);
      return data;
    },
    enabled: !!uploadRecordId,
  });

  const { data: templates } = useQuery<ContractTemplate[]>({
    queryKey: ["contract-templates"],
    queryFn: async () => {
      const { data } = await api.get("/contract-templates");
      return data;
    },
  });

  const { data: invoices } = useQuery<InvoiceRequest[]>({
    queryKey: ["invoice-requests"],
    queryFn: async () => {
      const { data } = await api.get("/invoice-requests");
      return data;
    },
  });

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("recordType", uploadRecordType);
      formData.append("recordId", uploadRecordId);
      const { data } = await api.post("/files/upload", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["files"] });
    },
  });

  const generateContractMutation = useMutation({
    mutationFn: async (templateId: string) => {
      const { data } = await api.post(`/contracts/generate`, { templateId });
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["files"] });
    },
  });

  const requestInvoiceMutation = useMutation({
    mutationFn: async (payload: { orderType: string; orderId: string; notes: string }) => {
      const { data } = await api.post("/invoice-requests", payload);
      return data;
    },
    onSuccess: () => {
      setShowInvoiceModal(false);
      setInvoiceOrderType("booking");
      setInvoiceOrderId("");
      setInvoiceNotes("");
      qc.invalidateQueries({ queryKey: ["invoice-requests"] });
    },
  });

  const handleDownload = async (fileId: string, fileName: string) => {
    try {
      const { data: tokenData } = await api.post(`/files/${fileId}/download-token`);
      const token = tokenData.token;
      const response = await api.get(`/files/download/${token}`, {
        responseType: "blob",
      });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement("a");
      link.href = url;
      link.download = fileName;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch {
    }
  };

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const fileList = e.target.files;
    if (fileList) {
      Array.from(fileList).forEach((file) => uploadMutation.mutate(file));
    }
  };

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const droppedFiles = e.dataTransfer.files;
      Array.from(droppedFiles).forEach((file) => uploadMutation.mutate(file));
    },
    [uploadMutation],
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setDragOver(false);
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Documents</h1>
        <p className="mt-1 text-gray-500">Manage and share documents.</p>
      </div>

      <div
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed p-8 transition-colors ${
          dragOver
            ? "border-indigo-400 bg-indigo-50"
            : "border-gray-300 bg-white hover:border-gray-400"
        }`}
      >
        <div className="mb-4 flex gap-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-600">Record Type</label>
            <select
              value={uploadRecordType}
              onChange={(e) => setUploadRecordType(e.target.value)}
              className="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
            >
              <option value="booking">Booking</option>
              <option value="procurement">Procurement</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-600">Record ID</label>
            <input
              type="text"
              value={uploadRecordId}
              onChange={(e) => setUploadRecordId(e.target.value)}
              placeholder="Enter record ID"
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
            />
          </div>
        </div>
        <p className="text-sm text-gray-600">
          Drag and drop files here, or{" "}
          <label className="cursor-pointer font-medium text-indigo-600 hover:text-indigo-700">
            browse
            <input
              type="file"
              multiple
              onChange={handleFileInput}
              className="hidden"
            />
          </label>
        </p>
        {uploadMutation.isPending && (
          <p className="mt-2 text-sm text-indigo-600">Uploading...</p>
        )}
        {uploadMutation.isError && (
          <p className="mt-2 text-sm text-red-600">Upload failed. Please try again.</p>
        )}
        {uploadMutation.isSuccess && (
          <p className="mt-2 text-sm text-green-600">File uploaded successfully.</p>
        )}
      </div>

      <div className="rounded-xl border border-gray-200 bg-white">
        <div className="border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">Files</h2>
        </div>
        {filesLoading ? (
          <div className="p-6 text-center text-gray-500">Loading files...</div>
        ) : !files?.length ? (
          <div className="p-6 text-center text-gray-500">
            No files uploaded yet.
          </div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="border-b border-gray-200 bg-gray-50">
              <tr>
                <th className="px-6 py-3 font-medium text-gray-600">Name</th>
                <th className="px-6 py-3 font-medium text-gray-600">Type</th>
                <th className="px-6 py-3 font-medium text-gray-600">Size</th>
                <th className="px-6 py-3 font-medium text-gray-600">Security</th>
                <th className="px-6 py-3 font-medium text-gray-600">Uploaded</th>
                <th className="px-6 py-3 font-medium text-gray-600">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {files.map((file) => (
                <tr key={file.id} className="hover:bg-gray-50">
                  <td className="px-6 py-3 font-medium text-gray-900">
                    {file.originalFilename}
                  </td>
                  <td className="px-6 py-3 text-gray-600">{file.mimeType}</td>
                  <td className="px-6 py-3 text-gray-600">
                    {formatBytes(file.byteSize)}
                  </td>
                  <td className="px-6 py-3">
                    {file.encrypted ? (
                      <span className="rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-700">
                        Encrypted
                      </span>
                    ) : (
                      <span className="text-xs text-gray-400">Standard</span>
                    )}
                  </td>
                  <td className="px-6 py-3 text-gray-500">
                    {formatDate(file.createdAt)}
                  </td>
                  <td className="px-6 py-3">
                    <button
                      onClick={() => handleDownload(file.id, file.originalFilename)}
                      className="text-sm font-medium text-indigo-600 hover:text-indigo-700"
                    >
                      Download
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-6">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          Contract Templates
        </h2>
        {!templates?.length ? (
          <p className="text-sm text-gray-500">No contract templates available.</p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {templates.map((tmpl) => (
              <div
                key={tmpl.id}
                className="rounded-lg border border-gray-200 p-4"
              >
                <h3 className="text-sm font-medium text-gray-900">{tmpl.name}</h3>
                <p className="mt-1 text-xs text-gray-500">{tmpl.description}</p>
                <p className="mt-1 text-xs text-gray-400">v{tmpl.version}</p>
                <button
                  onClick={() => generateContractMutation.mutate(tmpl.id)}
                  disabled={generateContractMutation.isPending}
                  className="mt-3 rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
                >
                  Generate Contract
                </button>
              </div>
            ))}
          </div>
        )}
        {generateContractMutation.isSuccess && (
          <p className="mt-3 text-sm text-green-600">
            Contract generated successfully.
          </p>
        )}
      </div>

      <div className="rounded-xl border border-gray-200 bg-white">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Invoice Requests
          </h2>
          <button
            onClick={() => setShowInvoiceModal(true)}
            className="rounded-lg border border-indigo-300 px-3 py-1.5 text-sm font-medium text-indigo-600 transition-colors hover:bg-indigo-50"
          >
            Request Invoice
          </button>
        </div>
        {!invoices?.length ? (
          <div className="p-6 text-center text-gray-500">
            No invoice requests yet.
          </div>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="border-b border-gray-200 bg-gray-50">
              <tr>
                <th className="px-6 py-3 font-medium text-gray-600">
                  Order
                </th>
                <th className="px-6 py-3 font-medium text-gray-600">Status</th>
                <th className="px-6 py-3 font-medium text-gray-600">Notes</th>
                <th className="px-6 py-3 font-medium text-gray-600">Requested</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {invoices.map((inv) => (
                <tr key={inv.id} className="hover:bg-gray-50">
                  <td className="px-6 py-3 font-mono text-xs text-gray-700">
                    {inv.orderType}: {inv.orderId}
                  </td>
                  <td className="px-6 py-3">
                    <span
                      className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[inv.status] ?? "bg-gray-100 text-gray-700"}`}
                    >
                      {statusLabel(inv.status)}
                    </span>
                  </td>
                  <td className="px-6 py-3 text-gray-700">
                    {inv.notes || "--"}
                  </td>
                  <td className="px-6 py-3 text-gray-500">
                    {formatDate(inv.createdAt)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showInvoiceModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-gray-900">
              Request Invoice
            </h3>
            <div className="space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Order Type
                </label>
                <select
                  value={invoiceOrderType}
                  onChange={(e) => setInvoiceOrderType(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                >
                  <option value="booking">Booking</option>
                  <option value="procurement">Procurement</option>
                </select>
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Order ID
                </label>
                <input
                  type="text"
                  value={invoiceOrderId}
                  onChange={(e) => setInvoiceOrderId(e.target.value)}
                  placeholder="Enter order ID"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Notes
                </label>
                <textarea
                  value={invoiceNotes}
                  onChange={(e) => setInvoiceNotes(e.target.value)}
                  rows={2}
                  placeholder="Optional notes"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
                />
              </div>
              {requestInvoiceMutation.isError && (
                <p className="text-sm text-red-600">
                  Failed to request invoice. Please try again.
                </p>
              )}
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                onClick={() => setShowInvoiceModal(false)}
                className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
              >
                Cancel
              </button>
              <button
                onClick={() =>
                  requestInvoiceMutation.mutate({
                    orderType: invoiceOrderType,
                    orderId: invoiceOrderId,
                    notes: invoiceNotes,
                  })
                }
                disabled={
                  !invoiceOrderId ||
                  requestInvoiceMutation.isPending
                }
                className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
              >
                {requestInvoiceMutation.isPending
                  ? "Requesting..."
                  : "Submit Request"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
