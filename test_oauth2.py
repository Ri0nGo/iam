import argparse
import http.cookiejar
import json
import sys
import urllib.error
import urllib.parse
import urllib.request


class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def request_json(opener, method, url, payload=None):
    data = None
    headers = {"Accept": "application/json"}
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with opener.open(req, timeout=10) as resp:
            body = resp.read().decode("utf-8")
            return resp.status, json.loads(body) if body else None
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8")
        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            parsed = body
        return exc.code, parsed


def request_redirect(opener, url):
    req = urllib.request.Request(url, method="GET")
    try:
        with opener.open(req, timeout=10) as resp:
            return resp.status, resp.headers.get("Location")
    except urllib.error.HTTPError as exc:
        if exc.code in (301, 302, 303, 307, 308):
            return exc.code, exc.headers.get("Location")
        body = exc.read().decode("utf-8")
        try:
            parsed = json.loads(body)
        except json.JSONDecodeError:
            parsed = body
        print_step("authorize failed", exc.code, parsed)
        return exc.code, None


def unwrap(resp):
    if isinstance(resp, dict) and "data" in resp:
        return resp["data"]
    return resp


def print_step(title, status, body):
    print(f"\n== {title} ==")
    print(f"HTTP {status}")
    print(json.dumps(body, ensure_ascii=False, indent=2))


def main():
    parser = argparse.ArgumentParser(description="Test IAM OAuth2 authorization code flow")
    parser.add_argument("--base-url", default="http://localhost:5173/api/v1")
    parser.add_argument("--client-id", required=True)
    parser.add_argument("--client-secret", required=True)
    parser.add_argument("--redirect-uri", required=True)
    parser.add_argument("--username", default="admin")
    parser.add_argument("--password", default="123456")
    parser.add_argument("--scope", default="basic")
    parser.add_argument("--state", default="xyz")
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/")
    cookie_jar = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cookie_jar), NoRedirectHandler())
    authorize_query = urllib.parse.urlencode({
        "response_type": "code",
        "client_id": args.client_id,
        "redirect_uri": args.redirect_uri,
        "scope": args.scope,
        "state": args.state,
    })

    login_payload = {"username": args.username, "password": args.password}
    status, login_resp = request_json(opener, "POST", f"{base_url}/auth/login", login_payload)
    print_step("1. login", status, login_resp)
    if status != 200:
        return 1

    status, location = request_redirect(opener, f"{base_url}/oauth/authorize?{authorize_query}")
    print(f"\n== 2. authorize redirect ==\nHTTP {status}\nLocation: {location}")
    if status != 302 or not location:
        return 1

    code = urllib.parse.parse_qs(urllib.parse.urlparse(location).query).get("code", [""])[0]
    if not code:
        print("No authorization code returned", file=sys.stderr)
        return 1

    token_payload = {
        "grant_type": "authorization_code",
        "client_id": args.client_id,
        "client_secret": args.client_secret,
        "code": code,
        "redirect_uri": args.redirect_uri,
    }
    status, token_resp = request_json(opener, "POST", f"{base_url}/oauth/token", token_payload)
    print_step("3. exchange token", status, token_resp)
    if status != 200:
        return 1

    access_token = token_resp.get("access_token")
    if not access_token:
        print("No access_token returned", file=sys.stderr)
        return 1

    userinfo_url = f"{base_url}/oauth/userinfo?" + urllib.parse.urlencode({"access_token": access_token})
    status, userinfo = request_json(opener, "GET", userinfo_url)
    print_step("4. userinfo", status, userinfo)
    return 0 if status == 200 else 1


if __name__ == "__main__":
    raise SystemExit(main())
