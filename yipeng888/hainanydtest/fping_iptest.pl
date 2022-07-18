#!/usr/bin/perl -w

print "\n\n ---------- Welcome ---------- \n\n"; 

#$num = @ARGV;
$inputfile = $ARGV[0];
$outputfile = $ARGV[1];

open (FileIn,"<$inputfile")|| die ("Could not open file"); 
#open (FileOut,">$outputfile");
$countline=`cat $inputfile|wc -l`;

pipe(PIPE_READ,PIPE_WRITE) or die "Can't make pipe!\n";

unless ($pid = fork) # master process, put all records in inputfile into pipe
{        
	defined $pid or die "can't fork:$!\n";
	close(PIPE_READ);

	while ($readaline = <FileIn>) 
	{
		chomp($readaline);
		@field = split(/\s+/, $readaline);
		print PIPE_WRITE "$field[0]\n";

		if(eof){ last; }
	}        

	close(FileIn);
	exit;
}

$SIG{CHLD} = sub { waitpid($pid,0) };

close(PIPE_WRITE);

my $pipecount=0;
my $maxlines=($countline -$countline %30)/30;
my @lines=();

while($record=<PIPE_READ>) #read from pipe,creat 50 subprocesses to do the fping job
{
	$pipecount++;
	chomp($record);

	push(@lines,$record);

	unless ($maxlines != 0 && $pipecount % $maxlines) # read $madlines records,then creat a subprocess.
	{           
		if (fork())
		{         
		#	print "\nStarting Sub_Process:$PID\n";
			@lines=();                  
		}
		else
		{
			foreach $ipreal (@lines)
			{
#				@IPd = split(/\//,$ip);
#				$ipreal=$IPd[0];
#				`fping $ipreal -c 10 -p 5 -i 10 -q 2>&1|awk -F / 'IP="'$ipreal'" {if(\$5~/^0%/) print IP,\$8}' >> $outputfile`;
#`fping $ipreal -c 10 -p 5 -i 10 -q 2>&1|awk -F , '{if(NF==2) print \$2}'|awk -F / 'IP="'$ipreal'" {print IP,\$4}' >> $outputfile`;
				`fping $ipreal -c 20 -p 200 -i 10 -q 2>&1|awk -F '[ /,]' 'IP="'$ipreal'"  {if(NF==17) print IP,\$9,\$16}'|tr -d '%' >> $outputfile`;
			}

#			close (FileOut);

   			exit 1;
		}
	}

}

foreach $ipreal (@lines) # deal with the remain records
{
#	@IPd = split(/\//,$ip);
#        $ipreal=$IPd[0];
#	`fping $ipreal -c 10 -p 5 -i 10 -q 2>&1|awk -F / 'IP="'$ipreal'" {if(\$5~/^0%/) print IP,\$8}' >> $outputfile`;
#`fping $ipreal -c 10 -p 5 -i 10 -q 2>&1|awk -F , '{if(NF==2) print \$2}'|awk -F / 'IP="'$ipreal'" {print IP,\$4}' >> $outputfile`;
	`fping $ipreal -c 20 -p 400 -i 10 -q 2>&1|awk -F '[ /,]' 'IP="'$ipreal'"  {if(NF==17) print IP,\$9,\$16}' |tr -d '%' >> $outputfile`;
}

#close (FileOut);

print "=========finish=========\n";

