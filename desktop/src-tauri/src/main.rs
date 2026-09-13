// Release builds use the Windows GUI subsystem to avoid a second console window.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    desktop_lib::run()
}
