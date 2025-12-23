pkgname=commit-ai
pkgver=1.0.0
pkgrel=1
pkgdesc="Generate Git commit messages using Gemini with Gitmoji or Conventional Commits"
arch=('any')
url="https://github.com/jhowk14/commit-ai"
license=('MIT')
depends=('git' 'jq' 'curl')
source=("commit-ai.sh")
sha256sums=('SKIP')

package() {
  install -Dm755 commit-ai.sh \
    "$pkgdir/usr/bin/commit-ai"
    
}
