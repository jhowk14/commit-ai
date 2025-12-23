pkgname=commit-ai
pkgver=1.0.0
pkgrel=1
pkgdesc="Generate Git commit messages using Gemini with Gitmoji or Conventional Commits"
arch=('any')
url="https://jhowk14.github.io/commit-ai/"
license=('MIT')
depends=('git' 'jq' 'curl')

source=("$pkgname-$pkgver.tar.gz::https://github.com/jhowk14/commit-ai/archive/refs/tags/v$pkgver.tar.gz")
sha256sums=('SKIP')

package() {
  cd "$srcdir/$pkgname-$pkgver"

  install -Dm755 commit-ai.sh \
    "$pkgdir/usr/bin/commit-ai"
    
  install -Dm644 README.md \
    "$pkgdir/usr/share/doc/$pkgname/README.md"
}
