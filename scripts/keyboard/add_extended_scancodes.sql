-- Add extended scancode mappings (E0-prefixed keys)
-- These are Windows extended scancodes (base scancode + 256)

INSERT OR REPLACE INTO scancode_mapping (scancode, key_name, key_category, description) VALUES
-- Extended Enter and Control
(284, 'Numpad Enter', 'special', 'Numpad Enter key (extended)'),
(285, 'Right Ctrl', 'modifier', 'Right Control key (extended)'),

-- Multimedia keys
(288, 'Mute', 'media', 'Mute/Unmute audio'),
(302, 'Volume Down', 'media', 'Decrease volume'),
(304, 'Volume Up', 'media', 'Increase volume'),

-- Extended numpad and special keys
(309, 'Numpad /', 'special', 'Numpad divide (extended)'),
(310, 'Right Shift', 'modifier', 'Right Shift key (extended)'),
(311, 'Print Screen', 'special', 'Print Screen key (extended)'),
(312, 'Right Alt', 'modifier', 'Right Alt key (extended)'),

-- Extended lock keys
(325, 'Num Lock', 'special', 'Num Lock key (extended)'),
(326, 'Pause', 'special', 'Pause/Break key (extended)'),

-- Extended navigation keys (arrow cluster)
(327, 'Home', 'special', 'Home key (extended)'),
(328, 'Up', 'special', 'Up arrow (extended)'),
(329, 'Page Up', 'special', 'Page Up key (extended)'),
(331, 'Left', 'special', 'Left arrow (extended)'),
(333, 'Right', 'special', 'Right arrow (extended)'),
(335, 'End', 'special', 'End key (extended)'),
(336, 'Down', 'special', 'Down arrow (extended)'),
(337, 'Page Down', 'special', 'Page Down key (extended)'),
(338, 'Insert', 'special', 'Insert key (extended)'),
(339, 'Delete', 'special', 'Delete key (extended)'),

-- Extended Windows keys
(347, 'Left Win', 'modifier', 'Left Windows key (extended)'),
(349, 'Menu', 'special', 'Context menu key (extended)'),

-- Additional media/function keys
(355, 'Media Key', 'media', 'Media control key'),
(554, 'Media Key', 'media', 'Media control key'),
(822, 'Media Key', 'media', 'Media control key');
