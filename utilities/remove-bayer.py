#!/usr/bin/env python

# Import the astropy fits tools
from astropy.io import fits
import glob, os, sys
from pathlib import Path

# Assume the first argument passed is the folder with the
# fits files in it
directory = sys.argv[1]

print(f'Looking for files in {directory}')

# Create a folder to store the output
output_folder = f'{directory}/converted-to-mono'
Path(output_folder).mkdir(parents=True, exist_ok=True)

# Loop for all fit(s) files in directory
for file in glob.glob(f'{directory}/*.fit*'):
    
    # Determine filename
    filename = os.path.basename(file)
    
    print(f'Processing {filename}')
    
    # Parse file as fits file
    hdulist = fits.open(file)
    # The primary header will contain the BAYERPAT
    header = hdulist[0].header
    
    # Ensure the file actually has the header. If so, delete it and write out to disk
    if "BAYERPAT" in header:
        print('Deleting BAYERPAT')
        del header['BAYERPAT']
        rewritten_file = f'{output_folder}/mono_{filename}'
        print(f'Writing result to {rewritten_file}')
        hdulist.writeto(rewritten_file)

    hdulist.close()
