package main

const cssStyles = `
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f8fafc;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }

        .report-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem;
            border-radius: 12px;
            margin-bottom: 2rem;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }

        .report-header h1 {
            font-size: 2.5rem;
            margin-bottom: 1rem;
            font-weight: 700;
        }

        .automation-info h2 {
            font-size: 1.8rem;
            margin-bottom: 0.5rem;
            font-weight: 600;
        }

        .description {
            font-size: 1.1rem;
            opacity: 0.9;
            margin-bottom: 1rem;
        }

        .meta-info {
            display: flex;
            gap: 1rem;
            flex-wrap: wrap;
            font-size: 0.9rem;
        }

        .meta-info span {
            background: rgba(255, 255, 255, 0.2);
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            backdrop-filter: blur(10px);
        }

        .status {
            font-weight: 600;
            text-transform: uppercase;
        }

        .status-success { background: rgba(34, 197, 94, 0.3) !important; }
        .status-error { background: rgba(239, 68, 68, 0.3) !important; }
        .status-warning { background: rgba(245, 158, 11, 0.3) !important; }

        .summary-section {
            margin-bottom: 2rem;
        }

        .summary-section h3 {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: #1f2937;
        }

        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }

        .summary-card {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }

        .card-header {
            background: #f3f4f6;
            padding: 1rem;
            font-weight: 600;
            color: #374151;
            border-bottom: 1px solid #e5e7eb;
        }

        .card-content {
            padding: 1rem;
        }

        .metric {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.5rem 0;
            border-bottom: 1px solid #f3f4f6;
        }

        .metric:last-child {
            border-bottom: none;
        }

        .metric-label {
            color: #6b7280;
            font-weight: 500;
        }

        .metric-value {
            font-weight: 600;
            color: #1f2937;
        }

        .metric-value.success {
            color: #059669;
        }

        .metric-value.error {
            color: #dc2626;
        }

        .error-section {
            background: #fef2f2;
            border: 1px solid #fecaca;
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 2rem;
        }

        .error-section h3 {
            color: #dc2626;
            margin-bottom: 1rem;
        }

        .error-message {
            background: white;
            padding: 1rem;
            border-radius: 6px;
            border-left: 4px solid #dc2626;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9rem;
            color: #7f1d1d;
        }

        .steps-section, .actions-section, .output-files-section {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .steps-section h3, .actions-section h3, .output-files-section h3 {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: #1f2937;
        }

        .table-container {
            overflow-x: auto;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 1rem;
        }

        th, td {
            padding: 0.75rem;
            text-align: left;
            border-bottom: 1px solid #e5e7eb;
        }

        th {
            background: #f9fafb;
            font-weight: 600;
            color: #374151;
            position: sticky;
            top: 0;
        }

        tr:hover {
            background: #f9fafb;
        }

        .success-rate {
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-weight: 600;
            font-size: 0.875rem;
        }

        .rate-excellent { background: #d1fae5; color: #065f46; }
        .rate-good { background: #dbeafe; color: #1e40af; }
        .rate-warning { background: #fef3c7; color: #92400e; }
        .rate-poor { background: #fee2e2; color: #991b1b; }

        .error-cell {
            max-width: 200px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            color: #dc2626;
            font-size: 0.875rem;
        }

        code {
            background: #f3f4f6;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.875rem;
        }

        .files-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 1rem;
            margin-top: 1rem;
        }

        .file-item {
            border: 1px solid #e5e7eb;
            border-radius: 6px;
            padding: 1rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            transition: all 0.2s;
        }

        .file-item:hover {
            border-color: #3b82f6;
            box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
        }

        .file-icon {
            font-size: 2rem;
        }

        .file-info {
            flex: 1;
        }

        .file-name {
            font-weight: 600;
            color: #1f2937;
            margin-bottom: 0.25rem;
            word-break: break-all;
        }

        .file-type {
            font-size: 0.75rem;
            color: #6b7280;
            text-transform: uppercase;
            margin-bottom: 0.5rem;
        }

        .file-link {
            color: #3b82f6;
            text-decoration: none;
            font-size: 0.875rem;
            font-weight: 500;
        }

        .file-link:hover {
            text-decoration: underline;
        }

        .no-files {
            text-align: center;
            color: #6b7280;
            font-style: italic;
            padding: 2rem;
        }

        .report-footer {
            text-align: center;
            padding: 2rem;
            color: #6b7280;
            font-size: 0.875rem;
            border-top: 1px solid #e5e7eb;
            margin-top: 2rem;
        }

        .report-footer a {
            color: #3b82f6;
            text-decoration: none;
        }

        .report-footer a:hover {
            text-decoration: underline;
        }

        @media (max-width: 768px) {
            .container {
                padding: 1rem;
            }
            
            .report-header {
                padding: 1.5rem;
            }
            
            .report-header h1 {
                font-size: 2rem;
            }
            
            .automation-info h2 {
                font-size: 1.5rem;
            }
            
            .summary-grid {
                grid-template-columns: 1fr;
            }
            
            .meta-info {
                flex-direction: column;
                gap: 0.5rem;
            }
            
            table {
                font-size: 0.875rem;
            }
            
            th, td {
                padding: 0.5rem;
            }
        }

        @media print {
            body {
                background: white;
            }
            
            .container {
                max-width: none;
                padding: 0;
            }
            
            .report-header {
                background: #667eea !important;
                -webkit-print-color-adjust: exact;
                color-adjust: exact;
            }
            
            .file-item:hover {
                border-color: #e5e7eb;
                box-shadow: none;
            }
        }
    `