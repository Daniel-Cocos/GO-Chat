import sys


def caesar_encrypt(text, shift=3):
    result = []
    for char in text:
        if char.isalpha():
            base = ord("A") if char.isupper() else ord("a")
            result.append(chr((ord(char) - base + shift) % 26 + base))
        else:
            result.append(char)
    return "".join(result)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Missing input text")
        sys.exit(1)

    input_text = sys.argv[1]
    encrypted = caesar_encrypt(input_text)
    print(encrypted)
