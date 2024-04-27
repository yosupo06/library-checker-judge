import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableRow from "@mui/material/TableRow";
import TableHead from "@mui/material/TableHead";

const JudgeStatusList = (): JSX.Element => {
  const judge_status = [
    {
      name: "AC",
      text: "Accepted (Green check: with the latest testcases)",
    },
    {
      name: "WA",
      text: "Wrong Answer",
    },
    {
      name: "RE",
      text: "Runtime Error",
    },
    {
      name: "TLE",
      text: "Time Limit Exceeded",
    },
    {
      name: "PE",
      text: "Presentation Error",
    },
    {
      name: "Fail",
      text: "The author's solution is wrong",
    },
    {
      name: "CE",
      text: "Compile Error",
    },
    {
      name: "WJ",
      text: "Waiting Judge",
    },
    {
      name: "IE",
      text: "Judge Server is broken 😢, please report this to the admin",
    },
    {
      name: "ICE",
      text: "Internal Compiler Error 😢, please report this to the admin",
    },
  ];

  return (
    <TableContainer component={Paper}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Status</TableCell>
            <TableCell>Info</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {judge_status.map((row) => (
            <TableRow key={row.name}>
              <TableCell>{row.name}</TableCell>
              <TableCell>{row.text}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default JudgeStatusList;
