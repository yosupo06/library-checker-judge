import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Grid from "@mui/material/Grid";
import React from "react";
import { useMonitoring } from "../api/client_wrapper";
import CircularProgress from "@mui/material/CircularProgress";
import Alert from "@mui/material/Alert";

const MonitoringPage: React.FC = () => {
  const monitoringQuery = useMonitoring();

  if (monitoringQuery.isPending) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" p={4}>
        <CircularProgress />
      </Box>
    );
  }

  if (monitoringQuery.isError) {
    return (
      <Box p={4}>
        <Alert severity="error">
          Error loading monitoring data: {monitoringQuery.error.message}
        </Alert>
      </Box>
    );
  }

  const data = monitoringQuery.data;

  return (
    <Box p={4}>
      <Typography variant="h4" gutterBottom>
        System Monitoring
      </Typography>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 3 }}>
            <Typography variant="h6" color="primary" gutterBottom>
              User Statistics
            </Typography>
            <Box display="flex" flexDirection="column" gap={1}>
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body1">Total Users:</Typography>
                <Typography variant="h6" color="text.primary">
                  {data.totalUsers.toLocaleString()}
                </Typography>
              </Box>
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body1">Total Submissions:</Typography>
                <Typography variant="h6" color="text.primary">
                  {data.totalSubmissions.toLocaleString()}
                </Typography>
              </Box>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper elevation={2} sx={{ p: 3 }}>
            <Typography variant="h6" color="primary" gutterBottom>
              Judge Queue Status
            </Typography>
            <Box display="flex" flexDirection="column" gap={1}>
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body1">Pending Tasks:</Typography>
                <Typography
                  variant="h6"
                  color={data.taskQueue?.pendingTasks > 10 ? "warning.main" : "text.primary"}
                >
                  {data.taskQueue?.pendingTasks ?? 0}
                </Typography>
              </Box>
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body1">Running Tasks:</Typography>
                <Typography variant="h6" color="success.main">
                  {data.taskQueue?.runningTasks ?? 0}
                </Typography>
              </Box>
              <Box display="flex" justifyContent="space-between">
                <Typography variant="body1">Total Tasks:</Typography>
                <Typography variant="h6" color="text.primary">
                  {data.taskQueue?.totalTasks ?? 0}
                </Typography>
              </Box>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12}>
          <Paper elevation={2} sx={{ p: 3 }}>
            <Typography variant="h6" color="primary" gutterBottom>
              System Health
            </Typography>
            <Box display="flex" flexDirection="column" gap={2}>
              <Box display="flex" alignItems="center" gap={2}>
                <Typography variant="body1">Queue Status:</Typography>
                {(data.taskQueue?.pendingTasks ?? 0) > 50 ? (
                  <Alert severity="warning" sx={{ py: 0 }}>
                    High load - {data.taskQueue?.pendingTasks} pending tasks
                  </Alert>
                ) : (data.taskQueue?.pendingTasks ?? 0) > 10 ? (
                  <Alert severity="info" sx={{ py: 0 }}>
                    Moderate load - {data.taskQueue?.pendingTasks} pending tasks
                  </Alert>
                ) : (
                  <Alert severity="success" sx={{ py: 0 }}>
                    Normal load - {data.taskQueue?.pendingTasks ?? 0} pending tasks
                  </Alert>
                )}
              </Box>
              <Typography variant="body2" color="text.secondary">
                Data refreshes automatically every 30 seconds
              </Typography>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default MonitoringPage;