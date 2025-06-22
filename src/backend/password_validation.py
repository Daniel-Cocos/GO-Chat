import string
import sys


def validate(pw: str):
    errors = []

    if len(pw) < 12:
        errors.append("Password must be at least 12 characters.")

    if not any(c.isupper() for c in pw):
        errors.append("Password must contain at least one uppercase letter.")

    if not any(c not in string.ascii_letters + string.digits for c in pw):
        errors.append("Password must contain at least one symbol.")

    return errors


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("No password provided.")
        sys.exit(1)

    password = sys.argv[1]
    errs = validate(password)
    if errs:
        for e in errs:
            print(e)
        sys.exit(1)

    sys.exit(0)
