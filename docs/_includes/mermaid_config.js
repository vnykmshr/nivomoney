// Mermaid Configuration - Theme-Aware Rendering
// Automatically switches between light and dark themes based on color scheme
{
  startOnLoad: false,
  theme: jtd.getTheme() === 'nivo-dark' ? 'dark' : 'default',
  themeVariables: jtd.getTheme() === 'nivo-dark' ? {
    // Dark theme variables matching nivo-dark.scss
    primaryColor: '#3b82f6',
    primaryTextColor: '#f9fafb',
    primaryBorderColor: '#60a5fa',
    lineColor: '#6b7280',
    secondaryColor: '#1f2937',
    tertiaryColor: '#374151',
    background: '#111827',
    mainBkg: '#1f2937',
    nodeBorder: '#4b5563',
    clusterBkg: '#1f2937',
    clusterBorder: '#4b5563',
    titleColor: '#f9fafb',
    edgeLabelBackground: '#1f2937',
    textColor: '#e5e7eb',
    nodeTextColor: '#f9fafb'
  } : {
    // Light theme variables matching nivo.scss
    primaryColor: '#3b82f6',
    primaryTextColor: '#ffffff',
    primaryBorderColor: '#2563eb',
    lineColor: '#6b7280',
    secondaryColor: '#f3f4f6',
    tertiaryColor: '#e5e7eb',
    background: '#ffffff',
    mainBkg: '#f9fafb',
    nodeBorder: '#d1d5db',
    clusterBkg: '#f3f4f6',
    clusterBorder: '#d1d5db',
    titleColor: '#111827',
    edgeLabelBackground: '#ffffff',
    textColor: '#374151',
    nodeTextColor: '#111827'
  },
  flowchart: {
    useMaxWidth: true,
    htmlLabels: true,
    curve: 'basis'
  },
  sequence: {
    useMaxWidth: true,
    diagramMarginX: 50,
    diagramMarginY: 10,
    actorMargin: 50,
    width: 150,
    height: 65,
    boxMargin: 10,
    boxTextMargin: 5,
    noteMargin: 10,
    messageMargin: 35,
    mirrorActors: true,
    bottomMarginAdj: 1,
    showSequenceNumbers: false
  }
}
