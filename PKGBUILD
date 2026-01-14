pkgname=commit-ai
pkgver=1.2.0
pkgrel=1
pkgdesc="Generate Git commit messages using Gemini with Gitmoji or Conventional Commits"
arch=('any')
url="https://jhowk14.github.io/commit-ai/"
license=('MIT')
depends=('git' 'jq' 'curl')

source=("commit-ai.sh::https://raw.githubusercontent.com/jhowk14/commit-ai/v${pkgver}/any-linux/commit-ai.sh")
sha256sums=('SKIP')

package() {
  install -Dm755 "$srcdir/commit-ai.sh" \
    "$pkgdir/usr/bin/commit-ai"
}
