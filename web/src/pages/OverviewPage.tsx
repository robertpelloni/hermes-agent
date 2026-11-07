import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  Activity,
  Brain,
  Clock,
  Cpu,
  FileText,
  FolderOpen,
  HelpCircle,
  KeyRound,
  MessageSquare,
  Package,
  Plug,
  Radio,
  RefreshCw,
  Settings,
  ShieldCheck,
  Terminal,
  Users,
  Webhook,
  Zap,
} from "lucide-react";
import { Button } from "@nous-research/ui/ui/components/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@nous-research/ui/ui/components/card";
import { Badge } from "@nous-research/ui/ui/components/badge";
import { Typography } from "@nous-research/ui/ui/components/typography/index";
import { Spinner } from "@nous-research/ui/ui/components/spinner";
import { api, type StatusResponse, type ModelOptionsResponse, type PaginatedSessions, type CronJob } from "@/lib/api";
import { usePageHeader } from "@/contexts/usePageHeader";

export default function OverviewPage() {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [modelOpts, setModelOpts] = useState<ModelOptionsResponse | null>(null);
  const [sessions, setSessions] = useState<PaginatedSessions | null>(null);
  const [cronJobs, setCronJobs] = useState<CronJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const { setEnd, setTitle } = usePageHeader();

  const loadData = useCallback(async () => {
    try {
      const [st, mo, sess, cj] = await Promise.allSettled([
        api.getStatus(),
        api.getModelOptions(),
        api.getSessions(5, 0),
        api.getCronJobs("all"),
      ]);

      if (st.status === "fulfilled") setStatus(st.value);
      if (mo.status === "fulfilled") setModelOpts(mo.value);
      if (sess.status === "fulfilled") setSessions(sess.value);
      if (cj.status === "fulfilled" && Array.isArray(cj.value)) setCronJobs(cj.value);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    setTitle("Command Center & Overview");
    loadData();
  }, [setTitle, loadData]);

  const handleRefresh = useCallback(() => {
    setRefreshing(true);
    loadData();
  }, [loadData]);

  useEffect(() => {
    setEnd(
      <Button
        ghost
        size="sm"
        onClick={handleRefresh}
        disabled={refreshing}
        className="gap-2"
        aria-label="Refresh live status metrics from Hermes Agent"
      >
        <RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
        <span>Refresh</span>
      </Button>
    );
  }, [handleRefresh, refreshing, setEnd]);

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center p-12">
        <Spinner className="h-8 w-8 text-primary" />
      </div>
    );
  }

  const activeCronCount = cronJobs.filter((j) => j.enabled || j.state === "active").length;
  const totalSessions = sessions?.total ?? 0;
  const currentModel = modelOpts?.model || "Hermes Agent Default";
  const currentProvider = modelOpts?.provider || "Configured Provider";
  const gatewayState = status?.gateway_running ? "Running" : (status?.gateway_state ?? "Ready");

  return (
    <div className="flex-1 space-y-6 overflow-y-auto p-6">
      {/* Top Banner / Hero Metric Strip */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Agent Telemetry Card */}
        <Card className="relative overflow-hidden border-border/40 bg-card/60 backdrop-blur">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Agent State
            </CardTitle>
            <div
              className="flex items-center text-muted-foreground hover:text-foreground cursor-help"
              title="Current runtime state of the agent server and background services"
            >
              <HelpCircle className="h-3.5 w-3.5" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <span className="flex h-2.5 w-2.5 rounded-full bg-emerald-500 shadow-sm" />
              <Typography className="text-2xl font-bold tracking-tight">Active</Typography>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              Hermes Gateway: {gatewayState}
            </p>
          </CardContent>
        </Card>

        {/* Active Model Card */}
        <Card className="relative overflow-hidden border-border/40 bg-card/60 backdrop-blur">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Active Model
            </CardTitle>
            <div
              className="flex items-center text-muted-foreground hover:text-foreground cursor-help"
              title="Primary reasoning and tool-calling model currently dispatched"
            >
              <Cpu className="h-4 w-4 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            <Typography className="truncate text-lg font-bold tracking-tight" title={currentModel}>
              {currentModel}
            </Typography>
            <p className="mt-1 text-xs text-muted-foreground">
              Provider: <span className="font-medium text-foreground">{currentProvider}</span>
            </p>
          </CardContent>
        </Card>

        {/* Sessions Count Card */}
        <Card className="relative overflow-hidden border-border/40 bg-card/60 backdrop-blur">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Total Sessions
            </CardTitle>
            <div
              className="flex items-center text-muted-foreground hover:text-foreground cursor-help"
              title="Total conversation sessions recorded in persistent memory"
            >
              <MessageSquare className="h-4 w-4 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            <Typography className="text-2xl font-bold tracking-tight">{totalSessions}</Typography>
            <p className="mt-1 text-xs text-muted-foreground">
              Recent: {sessions?.sessions?.length ?? 0} active loaded
            </p>
          </CardContent>
        </Card>

        {/* Scheduled Tasks Card */}
        <Card className="relative overflow-hidden border-border/40 bg-card/60 backdrop-blur">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Cron & Automations
            </CardTitle>
            <div
              className="flex items-center text-muted-foreground hover:text-foreground cursor-help"
              title="Active scheduled recurring tasks and automation blueprints"
            >
              <Clock className="h-4 w-4 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <Typography className="text-2xl font-bold tracking-tight">{activeCronCount}</Typography>
              <span className="text-xs text-muted-foreground">/ {cronJobs.length} total</span>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              {activeCronCount > 0 ? "Automations running on schedule" : "No scheduled tasks active"}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Consolidated Dashboard Hub */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Left Column (2 spans): Quick Action Workspace & Hub */}
        <div className="space-y-6 lg:col-span-2">
          {/* Quick Launch & Workflows */}
          <Card className="border-border/40 bg-card/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base font-semibold">
                <Zap className="h-4 w-4 text-amber-500" />
                Quick Workflows & Actions
              </CardTitle>
              <CardDescription>
                Direct access to high-value agent features and command launchers
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                <Link to="/chat" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <Terminal className="h-6 w-6 text-primary transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">Terminal Chat</span>
                    <span className="text-[11px] text-muted-foreground">Interactive PTY & TUI</span>
                  </div>
                </Link>

                <Link to="/sessions" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <MessageSquare className="h-6 w-6 text-blue-500 transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">Browse Sessions</span>
                    <span className="text-[11px] text-muted-foreground">Inspect history & trees</span>
                  </div>
                </Link>

                <Link to="/models" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <Brain className="h-6 w-6 text-purple-500 transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">Model Routing</span>
                    <span className="text-[11px] text-muted-foreground">Switch LLM & MoA</span>
                  </div>
                </Link>

                <Link to="/skills" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <Package className="h-6 w-6 text-emerald-500 transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">Skills & Packs</span>
                    <span className="text-[11px] text-muted-foreground">Configure tool bundles</span>
                  </div>
                </Link>

                <Link to="/mcp" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <Plug className="h-6 w-6 text-teal-500 transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">MCP Servers</span>
                    <span className="text-[11px] text-muted-foreground">Connect protocols</span>
                  </div>
                </Link>

                <Link to="/cron" className="block">
                  <div className="group flex flex-col items-center justify-center rounded-lg border border-border/40 bg-background/50 p-4 text-center transition-all hover:border-primary/50 hover:bg-muted/40">
                    <Clock className="h-6 w-6 text-amber-500 transition-transform group-hover:scale-110" />
                    <span className="mt-2 text-sm font-semibold">Cron & Tasks</span>
                    <span className="text-[11px] text-muted-foreground">Schedule automations</span>
                  </div>
                </Link>
              </div>
            </CardContent>
          </Card>

          {/* Consolidated Recent Activity */}
          <Card className="border-border/40 bg-card/50">
            <CardHeader className="flex flex-row items-center justify-between pb-3">
              <div>
                <CardTitle className="text-base font-semibold">Recent Sessions</CardTitle>
                <CardDescription>Latest agent conversations and tasks</CardDescription>
              </div>
              <Link to="/sessions">
                <Button ghost size="sm" className="text-xs">
                  View All &rarr;
                </Button>
              </Link>
            </CardHeader>
            <CardContent>
              {sessions?.sessions && sessions.sessions.length > 0 ? (
                <div className="divide-y divide-border/30">
                  {sessions.sessions.slice(0, 5).map((s) => (
                    <div key={s.id} className="flex items-center justify-between py-2.5">
                      <div className="min-w-0 flex-1 pr-4">
                        <Link
                          to={`/sessions?id=${encodeURIComponent(s.id)}`}
                          className="truncate text-sm font-medium hover:text-primary transition-colors block"
                        >
                          {s.title || "Untitled Session"}
                        </Link>
                        <p className="text-xs text-muted-foreground truncate">
                          ID: {s.id.slice(0, 8)} • {s.message_count ?? 0} messages •{" "}
                          {new Date(s.started_at || Date.now()).toLocaleDateString()}
                        </p>
                      </div>
                      <Badge className="text-[10px] shrink-0">
                        {s.model || "default"}
                      </Badge>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="py-6 text-center text-sm text-muted-foreground">
                  No recent sessions found. Start a conversation in Chat or via the CLI!
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Right Column: High-Value System & Capabilities Overview */}
        <div className="space-y-6">
          {/* System & Connectors Health */}
          <Card className="border-border/40 bg-card/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base font-semibold">
                <Activity className="h-4 w-4 text-primary" />
                Connectors & Gateways
              </CardTitle>
              <CardDescription>Live external platform status</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between rounded-md border border-border/30 bg-background/40 p-2.5">
                <div className="flex items-center gap-2">
                  <Radio className="h-4 w-4 text-blue-400" />
                  <span className="text-sm font-medium">Chat Gateways</span>
                </div>
                <Link to="/channels">
                  <Badge className="text-xs hover:bg-muted cursor-pointer">
                    Configure
                  </Badge>
                </Link>
              </div>

              <div className="flex items-center justify-between rounded-md border border-border/30 bg-background/40 p-2.5">
                <div className="flex items-center gap-2">
                  <Webhook className="h-4 w-4 text-amber-400" />
                  <span className="text-sm font-medium">Webhooks</span>
                </div>
                <Link to="/webhooks">
                  <Badge className="text-xs hover:bg-muted cursor-pointer">
                    Inspect
                  </Badge>
                </Link>
              </div>

              <div className="flex items-center justify-between rounded-md border border-border/30 bg-background/40 p-2.5">
                <div className="flex items-center gap-2">
                  <ShieldCheck className="h-4 w-4 text-emerald-400" />
                  <span className="text-sm font-medium">Device Pairing</span>
                </div>
                <Link to="/pairing">
                  <Badge className="text-xs hover:bg-muted cursor-pointer">
                    Manage
                  </Badge>
                </Link>
              </div>
            </CardContent>
          </Card>

          {/* Quick Config & Resources */}
          <Card className="border-border/40 bg-card/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base font-semibold">
                <Settings className="h-4 w-4 text-muted-foreground" />
                System & Configuration
              </CardTitle>
              <CardDescription>Security, profiles and system tools</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <Link
                to="/env"
                className="flex items-center justify-between rounded-md p-2 text-sm transition-colors hover:bg-muted/50"
              >
                <div className="flex items-center gap-2">
                  <KeyRound className="h-4 w-4 text-primary" />
                  <span>API Keys & Secrets</span>
                </div>
                <span className="text-xs text-muted-foreground">&rarr;</span>
              </Link>

              <Link
                to="/profiles"
                className="flex items-center justify-between rounded-md p-2 text-sm transition-colors hover:bg-muted/50"
              >
                <div className="flex items-center gap-2">
                  <Users className="h-4 w-4 text-blue-400" />
                  <span>Agent Profiles</span>
                </div>
                <span className="text-xs text-muted-foreground">&rarr;</span>
              </Link>

              <Link
                to="/files"
                className="flex items-center justify-between rounded-md p-2 text-sm transition-colors hover:bg-muted/50"
              >
                <div className="flex items-center gap-2">
                  <FolderOpen className="h-4 w-4 text-emerald-400" />
                  <span>Workspace File Browser</span>
                </div>
                <span className="text-xs text-muted-foreground">&rarr;</span>
              </Link>

              <Link
                to="/logs"
                className="flex items-center justify-between rounded-md p-2 text-sm transition-colors hover:bg-muted/50"
              >
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-amber-400" />
                  <span>System Diagnostics & Logs</span>
                </div>
                <span className="text-xs text-muted-foreground">&rarr;</span>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
