import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "30s",
};

export function setup() {
  const loginRes = http.post(
    "http://localhost:8080/login",
    JSON.stringify({
      username: "employee1",
      password: "password123",
    }),
    {
      headers: { "Content-Type": "application/json" },
    }
  );

  const token = JSON.parse(loginRes.body).token;
  return { token };
}

export default function (data) {
  const headers = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${data.token}`,
  };

  // 1. Check-in
  const checkin = http.post(
    "http://localhost:8080/api/attendance/checkin",
    JSON.stringify({
      user_id: 1,
      attendance_period_id: 1,
    }),
    { headers }
  );
  check(checkin, { "check-in success": (r) => r.status === 200 });
  sleep(1);

  // 2. Check-out
  const checkout = http.patch(
    "http://localhost:8080/api/attendance/checkout",
    JSON.stringify({
      user_id: 1,
      attendance_period_id: 1,
      check_out_at: new Date().toISOString(),
    }),
    { headers }
  );
  check(checkout, {
    "check-out success": (r) => r.status === 200 || r.status === 409,
  });
  sleep(1);

  // 3. Request Overtime
  const overtime = http.post(
    "http://localhost:8080/api/attendance/overtime",
    JSON.stringify({
      user_id: 1,
      attendance_period_id: 1,
      date: new Date().toISOString().split("T")[0],
      hours: 2,
    }),
    { headers }
  );
  check(overtime, { "overtime requested": (r) => r.status === 200 });
  sleep(1);

  // 4. Submit Reimbursement
  const reimbursement = http.post(
    "http://localhost:8080/api/reimbursement/submit",
    JSON.stringify({
      user_id: 1,
      attendance_period_id: 1,
      title: "Meal Allowance",
      amount: 30000,
    }),
    { headers }
  );
  check(reimbursement, { "reimbursement submitted": (r) => r.status === 200 });
  sleep(1);
}
