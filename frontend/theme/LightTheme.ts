import type { ThemeTypes } from '@/types/themeTypes/ThemeType';

// Modern slate + indigo palette. Key names are kept so existing pages
// (bg-lightprimary, text-textPrimary, ...) continue to work.
const BLUE_THEME: ThemeTypes = {
    name: 'BLUE_THEME',
    dark: false,
    variables: {
        'border-color': '#e6e8f0'
    },
    colors: {
        primary: '#4f46e5',
        secondary: '#64748b',
        info: '#0ea5e9',
        success: '#10b981',
        warning: '#f59e0b',
        error: '#ef4444',
        indigo: '#7c3aed',
        lightprimary: '#eef2ff',
        lightinfo: '#e0f2fe',
        lightsecondary: '#f1f5f9',
        lightsuccess: '#d1fae5',
        lighterror: '#fee2e2',
        lightwarning: '#fef3c7',
        lightindigo: '#ede9fe',
        textPrimary: '#0f172a',
        textSecondary: '#475569',
        borderColor: '#e6e8f0',
        inputBorder: '#d9dce6',
        containerBg: '#ffffff',
        background: '#f6f7fb',
        hoverColor: '#f8fafc',
        surface: '#ffffff',
        grey100: '#64748b',
        grey200: '#0f172a',
        darkbg: '#1e293b',
        bglight: '#f8fafc',
        bgdark: '#0f172a'
    }
};

export { BLUE_THEME };
