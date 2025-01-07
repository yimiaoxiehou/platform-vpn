//go:build !android || cmfa

package sing_tun

import (
	tun "github.com/metacubex/sing-tun"
)

func (l *Listener) buildAndroidRules(_ *tun.Options) error {
	return nil
}
func (l *Listener) openAndroidHotspot(_ tun.Options) {}
