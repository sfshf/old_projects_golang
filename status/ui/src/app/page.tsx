"use client";
import Backdrop from "@mui/material/Backdrop";
import React, { useEffect } from "react";
import CircularProgress from "@mui/material/CircularProgress";
import Button from "@mui/material/Button";
import Snackbar, { SnackbarOrigin } from "@mui/material/Snackbar";
import MuiAlert, { AlertProps } from "@mui/material/Alert";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { request, IncomingMessage, RequestOptions } from "http";
import Box from "@mui/material/Box";
import Collapse from "@mui/material/Collapse";
import IconButton from "@mui/material/IconButton";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import moment from "moment";
import { green, red } from "@mui/material/colors";

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(function Alert(
  props,
  ref
) {
  return <MuiAlert elevation={6} ref={ref} variant="filled" {...props} />;
});

const localAddress = "https://api.test.n1xt.net";

const Row = ({ srvNow, row }: { srvNow: number; row: Status }) => {
  const [open, setOpen] = React.useState(false);

  return (
    <React.Fragment>
      <TableRow sx={{ "& > *": { borderBottom: "unset" } }}>
        <TableCell>
          <IconButton
            aria-label="expand row"
            size="small"
            onClick={() => setOpen(!open)}
          >
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell component="th" scope="row">
          {row.name}
        </TableCell>
        <TableCell>
          {((row.proportion / row.endpoints.length) * 100).toFixed(0)}%
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              <Typography variant="h6" gutterBottom component="div">
                Endpoints
              </Typography>
              <Table size="small" aria-label="purchases">
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Address</TableCell>
                    <TableCell>Method</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>LastTime</TableCell>
                    <TableCell>Duration</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {row.endpoints &&
                    row.endpoints.map((endpoint) => (
                      <TableRow key={endpoint.name}>
                        <TableCell component="th" scope="row">
                          {endpoint.name}
                        </TableCell>
                        <TableCell>
                          {endpoint.hostname + endpoint.path}
                        </TableCell>
                        <TableCell>{endpoint.method}</TableCell>
                        <TableCell
                          sx={
                            endpoint.status
                              ? { color: "#00ff00" }
                              : { color: "#ff0000" }
                          }
                        >
                          {endpoint.status ? "Good" : "Down"}
                        </TableCell>
                        <TableCell>
                          {moment
                            .unix(endpoint.lastDT)
                            .local()
                            .format("YYYY-MM-DD HH:mm:ss")}
                        </TableCell>
                        <TableCell>
                          {Math.trunc(
                            moment
                              .duration(
                                moment
                                  .unix(srvNow)
                                  .diff(moment.unix(endpoint.lastDT))
                              )
                              .asMinutes() / 60
                          ) +
                            ":" +
                            Math.round(
                              moment
                                .duration(
                                  moment
                                    .unix(srvNow)
                                    .diff(moment.unix(endpoint.lastDT))
                                )
                                .asMinutes() % 60
                            )}
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
};

interface Endpoint {
  name: string;
  hostname: string;
  path: string;
  method: string;
  status: boolean;
  param?: any;
  respCode?: any;
  respDataCode?: any;
  lastDT: number;
  currentDT: number;
  duration: string;
}

interface Status {
  name: string;
  proportion: number;
  endpoints: Endpoint[];
}

export default function Home() {
  const [loading, setLoading] = React.useState(false);
  const [toast, setToast] = React.useState("");
  const [errorToast, setErrorToast] = React.useState("");

  const [srvNow, setSrvNow] = React.useState(0);
  const [allStatus, setAllStatus] = React.useState<Status[] | null>(null);

  const handleRequest = async (
    url: string,
    method: string,
    reqData: any,
    respHandle?: (resp: IncomingMessage) => void,
    respDataHandle?: (chunk: any) => void
  ) => {
    setLoading(true);
    try {
      const reqOpts: RequestOptions = {
        path: url,
        method: method,
      };
      let postData = null;
      if (reqData) {
        postData = JSON.stringify(reqData);
        reqOpts.headers = {
          "Content-Type": "application/json",
          "Content-Length": Buffer.byteLength(postData),
        };
      }
      const req = request(reqOpts, (res) => {
        res.setEncoding("utf8");
        if (respDataHandle) {
          res.on("data", respDataHandle);
        }
      });
      if (respHandle) {
        req.on("response", respHandle);
      }
      req.on("error", (e) => {
        throw e;
      });
      if (postData) {
        req.write(postData);
      }
      req.end();

      setLoading(false);
    } catch (error) {
      const message = (error as Error).message;
      setLoading(false);
      setToast(message);
    }
  };

  useEffect(() => {
    handleRequest(
      localAddress + "/status/all",
      "POST",
      null,
      undefined,
      (chunk: any): void => {
        const respData = JSON.parse(chunk);
        setSrvNow(respData.now);
        setAllStatus(respData.list);
      }
    );
  }, []);

  return (
    <main>
      <TableContainer component={Paper}>
        <Table aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell />
              <TableCell>Service Name</TableCell>
              <TableCell>Proportion</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {allStatus &&
              allStatus.map((status) => {
                return <Row key={status.name} srvNow={srvNow} row={status} />;
              })}
          </TableBody>
        </Table>
      </TableContainer>

      <Snackbar
        open={errorToast !== ""}
        autoHideDuration={5000}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
        onClose={() => setErrorToast("")}
      >
        <Alert
          onClose={() => setErrorToast("")}
          severity="error"
          sx={{ width: "100%" }}
        >
          {errorToast}
        </Alert>
      </Snackbar>
      <Snackbar
        open={toast !== ""}
        autoHideDuration={4500}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
        onClose={() => setToast("")}
      >
        <Alert
          onClose={() => setToast("")}
          severity="error"
          sx={{ width: "100%" }}
        >
          {toast}
        </Alert>
      </Snackbar>
      <Backdrop
        sx={{ color: "#fff", zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
        // onClick={handleClose}
      >
        <CircularProgress color="inherit" />
      </Backdrop>
    </main>
  );
}
