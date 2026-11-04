// @vitest-environment jsdom
import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { BrowserRouter } from "react-router-dom";
import OverviewPage from "./OverviewPage";
import { PageHeaderProvider } from "@/contexts/PageHeaderProvider";

vi.mock("@/lib/api", () => ({
  api: {
    getStatus: vi.fn().mockResolvedValue({
      status: "running",
      gateway_running: true,
      gateway_state: "ready",
      version: "1.0.0",
    }),
    getModelOptions: vi.fn().mockResolvedValue({
      model: "hermes-3-llama-3.1-70b",
      provider: "openrouter",
    }),
    getSessions: vi.fn().mockResolvedValue({
      total: 12,
      sessions: [
        {
          id: "sess-1",
          title: "Code Refactoring",
          message_count: 5,
          started_at: Date.now(),
          model: "hermes-3",
        },
      ],
    }),
    getCronJobs: vi.fn().mockResolvedValue([
      { id: "cron-1", name: "Backup", enabled: true, state: "active" },
    ]),
  },
}));

describe("OverviewPage", () => {
  it("renders metrics, actions, and recent activity properly", async () => {
    render(
      <BrowserRouter>
        <PageHeaderProvider pluginTabs={[]}>
          <OverviewPage />
        </PageHeaderProvider>
      </BrowserRouter>
    );

    await waitFor(() => {
      expect(screen.getByText("Active Model")).toBeDefined();
    });

    expect(screen.getByText("Agent State")).toBeDefined();
    expect(screen.getByText("Total Sessions")).toBeDefined();
    expect(screen.getByText("Cron & Automations")).toBeDefined();
    expect(screen.getByText("Quick Workflows & Actions")).toBeDefined();
    expect(screen.getByText("Code Refactoring")).toBeDefined();
  });
});
